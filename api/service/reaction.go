/*
 * @Author: flwfdd
 * @Date: 2025-03-09 23:49:47
 * @LastEditTime: 2025-03-11 16:18:14
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/cache"
	"BIT101-GO/pkg/gorse"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zhenghaoz/gorse/client"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 检查是否实现接口
var _ types.ReactionService = (*ReactionService)(nil)

// ReactionService 点赞服务
type ReactionService struct {
	UserSvc    types.UserService
	ImageSvc   types.ImageService
	MessageSvc types.MessageService
}

// NewReactionService 创建点赞服务
func NewReactionService(userSvc types.UserService, imageSvc types.ImageService, messageSvc types.MessageService) *ReactionService {
	s := &ReactionService{
		UserSvc:    userSvc,
		ImageSvc:   imageSvc,
		MessageSvc: messageSvc,
	}
	types.RegisterObjHandler(s)
	return s
}

/* ObjHandler */

// GetObjType 获取对象类型
func (s *ReactionService) GetObjType() string {
	return "comment"
}

// IsExist 对象是否存在
func (s *ReactionService) IsExist(id uint) bool {
	if _, err := s.getComment(id); err != nil {
		return false
	}
	return true
}

// GetUid 获取对象的用户id
func (s *ReactionService) GetUid(id uint) (uint, error) {
	comment, err := s.getComment(id)
	if err != nil {
		return 0, err
	}
	return comment.Uid, nil
}

// LikeHandler 点赞处理
func (s *ReactionService) LikeHandler(tx *gorm.DB, id uint, delta int, fromUid uint) (likeNum uint, err error) {
	var comment database.Comment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&comment, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if comment.ID == 0 {
		return 0, errors.New("评论不存在Orz")
	}
	comment.LikeNum = uint(int(comment.LikeNum) + delta)
	if err := tx.Save(&comment).Error; err != nil {
		return 0, err
	}

	// 通知
	if delta == 1 && fromUid != comment.Uid {
		go func() {
			// 获取顶级对象 处理子评论问题
			obj, err := types.NewObj(comment.Obj)
			if err != nil {
				return
			}
			linkObj, err := s.getSuperObj(obj)
			if err != nil {
				return
			}

			s.MessageSvc.Send(int(fromUid), comment.Uid, obj.GetObjID(), types.MessageTypeLike, linkObj.GetObjID(), comment.Text)
		}()
	}

	return comment.LikeNum, nil
}

// CommentHandler 评论处理
func (s *ReactionService) CommentHandler(tx *gorm.DB, id uint, subComment database.Comment, delta int, fromUid uint) (commentNum uint, err error) {
	var comment database.Comment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&comment, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if comment.ID == 0 {
		return 0, errors.New("评论不存在Orz")
	}
	comment.CommentNum = uint(int(comment.CommentNum) + delta)
	if err := tx.Save(&comment).Error; err != nil {
		return 0, err
	}

	// 通知
	if delta == 1 {
		go func() {
			toUid := comment.Uid
			// 回复评论则通知回复用户
			replyObj, err := types.NewObj(subComment.ReplyObj)
			if err == nil {
				toUid_, err := replyObj.GetUid()
				if err == nil {
					toUid = toUid_
				}
			}
			if toUid == fromUid {
				return
			}

			// 获取顶级对象 处理子评论问题
			commentObj, err := types.NewObj(subComment.Obj)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			linkObj, err := s.getSuperObj(commentObj)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			fromUid := int(subComment.Uid)
			if subComment.Anonymous {
				fromUid = -1
			}
			s.MessageSvc.Send(fromUid, toUid, commentObj.GetObjID(), types.MessageTypeComment, linkObj.GetObjID(), subComment.Text)
		}()
	}

	return comment.CommentNum, nil
}

// Like 点赞
func (s *ReactionService) Like(obj types.Obj, uid uint) (types.LikeAPI, error) {
	// 限流
	set, err := cache.RDB.SetNX(cache.Context, fmt.Sprintf("like:%d_%s", uid, obj.GetObjID()), "1", time.Second).Result()
	if err != nil || !set {
		return types.LikeAPI{}, errors.New("操作过于频繁Orz")
	}

	var likeAPI types.LikeAPI
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		delta := 0
		var like database.Like
		tx.Unscoped().Where("uid = ?", uid).Where("obj = ?", obj.GetObjID()).Limit(1).Find(&like)
		if like.ID == 0 { //新建
			like = database.Like{
				Obj: obj.GetObjID(),
				Uid: uid,
			}
			if err := tx.Create(&like).Error; err != nil {
				return err
			}
			delta = 1
		} else if like.DeletedAt.Valid { //删除过 取消删除
			like.DeletedAt = gorm.DeletedAt{}
			if err := tx.Unscoped().Save(&like).Error; err != nil {
				return err
			}
			delta = 1

		} else { //取消点赞
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			delta = -1
		}

		// 更新对象点赞数
		likeAPI.LikeNum, err = obj.LikeHandler(tx, delta, uid)
		if err != nil {
			return err
		}
		likeAPI.Like = !like.DeletedAt.Valid
		return nil
	}); err != nil {
		return types.LikeAPI{}, err
	}
	return likeAPI, nil
}

// CheckLike 检查是否点赞
func (s *ReactionService) CheckLike(objID string, uid uint) bool {
	var like database.Like
	database.DB.Where("uid = ?", uid).Where("obj = ?", objID).Limit(1).Find(&like)
	return like.ID != 0
}

// CheckLikeMap 批量检查是否点赞
func (s *ReactionService) CheckLikeMap(objIDMap map[string]bool, uid uint) (map[string]bool, error) {
	objList := make([]string, 0, len(objIDMap))
	for obj := range objIDMap {
		objList = append(objList, obj)
		objIDMap[obj] = false
	}
	var like_list []database.Like
	if err := database.DB.Where("uid = ?", uid).Where("obj IN ?", objList).Find(&like_list).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	for _, like := range like_list {
		objIDMap[like.Obj] = true
	}
	return objIDMap, nil
}

// getComment 获取评论
func (s *ReactionService) getComment(id uint) (database.Comment, error) {
	var comment database.Comment
	if err := database.DB.Limit(1).Find(&comment, "id = ?", id).Error; err != nil {
		return database.Comment{}, errors.New("数据库错误Orz")
	}
	if comment.ID == 0 {
		return database.Comment{}, errors.New("评论不存在Orz")
	}
	return comment, nil
}

// 将数据库格式评论转化为返回格式
func (s *ReactionService) commnet2CommentAPI(oldComment database.Comment, uid uint, admin bool, superObjID string) (types.CommentAPI, error) {
	commentAPI, err := s.comments2CommentAPIs([]database.Comment{oldComment}, uid, admin, superObjID)
	if err != nil {
		return types.CommentAPI{}, err
	}
	if len(commentAPI) == 0 {
		return types.CommentAPI{}, errors.New("评论不存在Orz")
	}
	return commentAPI[0], nil
}

// getUidsFromComments 统计评论中的用户
func (s *ReactionService) getUidsFromComments(comments []database.Comment) map[int]bool {
	uidMap := make(map[int]bool)
	for _, comment := range comments {
		// 匿名用户
		if comment.Anonymous {
			uidMap[-1] = true
		} else {
			uidMap[int(comment.Uid)] = true
		}
		// 回复用户
		if comment.ReplyUid < 0 {
			uidMap[-1] = true
			uidMap[-comment.ReplyUid] = true
		} else if comment.ReplyUid > 0 {
			uidMap[comment.ReplyUid] = true
		}
	}
	return uidMap
}

// getLikeObjsFromComments 统计评论中的点赞对象
func (s *ReactionService) getLikeObjsFromComments(comments []database.Comment) map[string]bool {
	likeMap := make(map[string]bool)
	for _, comment := range comments {
		likeMap["comment"+fmt.Sprint(comment.ID)] = true
	}
	return likeMap
}

// getUserInCommentAPI 获取评论中的用户信息 用于最后组装
func (s *ReactionService) getUserAPIInComment(comment database.Comment, users map[int]types.UserAPI, superObj string) (user, replyUser types.UserAPI) {
	if comment.Anonymous {
		user = users[-1]
		user.Nickname = s.UserSvc.GetAnonymousName(comment.Uid, superObj)
	} else {
		user = users[int(comment.Uid)]
	}
	if int(comment.ReplyUid) < 0 {
		replyUser = users[-1]
		replyUser.Nickname = s.UserSvc.GetAnonymousName(uint(-int(comment.ReplyUid)), superObj)
		comment.ReplyUid = -1
	} else if int(comment.ReplyUid) > 0 {
		replyUser = users[int(comment.ReplyUid)]
	}
	return user, replyUser
}

// 批量将数据库格式评论转化为返回格式
func (s *ReactionService) comments2CommentAPIs(oldComments []database.Comment, uid uint, admin bool, superObjID string) ([]types.CommentAPI, error) {
	comments := make([]types.CommentAPI, 0)

	// 统计需要查询的用户和点赞
	uidMap := s.getUidsFromComments(oldComments)
	likeMap := s.getLikeObjsFromComments(oldComments)
	subCommentObjList := make([]string, 0)
	subCommentMap := make(map[string][]types.CommentAPI)
	for _, oldComment := range oldComments {
		oldCommentObj := fmt.Sprintf("comment%d", oldComment.ID)
		// 子评论
		if oldComment.CommentNum > 0 {
			subCommentObjList = append(subCommentObjList, oldCommentObj)
		}
		subCommentMap[oldCommentObj] = make([]types.CommentAPI, 0)
	}

	// 查询子评论
	var subComments []database.Comment
	if err := database.DB.Raw(`SELECT * FROM (SELECT *,ROW_NUMBER() OVER (PARTITION BY "obj" ORDER BY "like_num" DESC) AS rn FROM comments WHERE "deleted_at" IS NULL AND obj IN ?) t WHERE rn<=?`, subCommentObjList, config.Get().CommentPreviewSize).Scan(&subComments).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	for uid := range s.getUidsFromComments(subComments) {
		uidMap[uid] = true
	}
	for lieObj := range s.getLikeObjsFromComments(subComments) {
		likeMap[lieObj] = true
	}

	// 批量获取用户和点赞
	users, err := s.UserSvc.GetUserAPIMap(uidMap)
	if err != nil {
		return nil, err
	}
	likes, err := s.CheckLikeMap(likeMap, uid)
	if err != nil {
		return nil, err
	}

	// 组装子评论
	for _, subComment := range subComments {
		user, replyUser := s.getUserAPIInComment(subComment, users, superObjID)
		subComment.Uid = 0
		subCommentMap[subComment.Obj] = append(subCommentMap[subComment.Obj], types.CommentAPI{
			Comment:   subComment,
			Like:      likes[fmt.Sprintf("comment%d", subComment.ID)],
			Own:       subComment.Uid == uid || admin,
			ReplyUser: replyUser,
			User:      user,
			Sub:       make([]types.CommentAPI, 0),
			Images:    s.ImageSvc.GetImageAPIList(common.Spilt(subComment.Images)),
		})
	}

	// 组装评论
	for _, oldComment := range oldComments {
		commentObj := fmt.Sprintf("comment%d", oldComment.ID)
		user, replyUser := s.getUserAPIInComment(oldComment, users, superObjID)
		oldComment.Uid = 0
		comment := types.CommentAPI{
			Comment:   oldComment,
			Like:      likes[commentObj],
			Own:       oldComment.Uid == uid || admin,
			ReplyUser: replyUser,
			User:      user,
			Images:    s.ImageSvc.GetImageAPIList(common.Spilt(oldComment.Images)),
		}

		comment.Sub = subCommentMap[commentObj]
		comments = append(comments, comment)
	}

	return comments, nil
}

func (s *ReactionService) getSuperObj(obj types.Obj) (types.Obj, error) {
	for obj.GetObjType() == "comment" {
		comment, err := s.getComment(obj.GetID())
		if err != nil {
			return types.Obj{}, errors.New("获取顶级对象失败Orz")
		}
		obj, err = types.NewObj(comment.Obj)
		if err != nil {
			return types.Obj{}, errors.New("获取顶级对象失败Orz")
		}
	}
	return obj, nil
}

// Comment 评论
func (s *ReactionService) Comment(obj types.Obj, text string, anonymous bool, replyUid int, replyObj types.Obj, rate uint, mids []string, uid uint, admin bool) (types.CommentAPI, error) {
	// 限流
	set, err := cache.RDB.SetNX(cache.Context, fmt.Sprintf("comment:%d", uid), "1", time.Second).Result()
	if err != nil || !set {
		return types.CommentAPI{}, errors.New("操作过于频繁Orz")
	}

	if !s.ImageSvc.CheckMids(mids) {
		return types.CommentAPI{}, errors.New("存在未上传成功的图片Orz")
	}

	var commentAPI types.CommentAPI
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 评价最多一次
		var comment database.Comment
		if rate != 0 {
			tx.Limit(1).Where("uid = ?", uid).Where("obj = ?", obj.GetObjID()).Find(&comment)
			if comment.ID != 0 {
				return errors.New("不能重复评价Orz")
			}
		}

		// 回复匿名用户
		if replyUid == -1 {
			replyObjUid, err := replyObj.GetUid()
			if err != nil {
				return errors.New("获取回复用户失败Orz")
			}
			replyUid = -int(replyObjUid)
		} else if replyUid > 0 {
			if _, err := s.UserSvc.GetUserAPI(replyUid); err != nil {
				return errors.New("获取回复用户失败Orz")
			}
		}

		// 评论
		comment = database.Comment{
			Obj:       obj.GetObjID(),
			Text:      text,
			Anonymous: anonymous,
			ReplyUid:  replyUid,
			ReplyObj:  replyObj.GetObjID(),
			Rate:      uint(rate),
			Uid:       uid,
			Images:    strings.Join(mids, " "),
		}
		if err := tx.Create(&comment).Error; err != nil {
			return errors.New("数据库错误Orz")
		}

		// 评论数+1
		if _, err := obj.CommentHandler(tx, comment, 1, uid); err != nil {
			return errors.New("更新评论数失败Orz")
		}

		superObj, err := s.getSuperObj(obj)
		if err != nil {
			return err
		}
		commentAPI, err = s.commnet2CommentAPI(comment, uid, admin, superObj.GetObjID())
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return types.CommentAPI{}, err
	}
	return commentAPI, nil
}

// GetComments 获取评论列表
func (s *ReactionService) GetComments(obj types.Obj, order string, page uint, uid uint, admin bool) ([]types.CommentAPI, error) {
	// 获取根对象
	superObj, err := s.getSuperObj(obj)
	if err != nil {
		return nil, err
	}
	var comments []database.Comment
	q := database.DB.Model(&database.Comment{}).Where("obj = ?", obj.GetObjID())
	// 排序
	switch order {
	case "like": //点赞数多的在前
		q = q.Order("like_num DESC")
	case "old": //发布时间早的在前
		q = q.Order("created_at")
	case "new": //发布时间晚的在前
		q = q.Order("created_at DESC")
	default: //默认 状态新的在前
		q = q.Order("updated_at DESC")
	}
	// 分页
	page_size := config.Get().CommentPageSize
	q = q.Offset(int(page * page_size)).Limit(int(page_size))
	if err := q.Find(&comments).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	commentAPIs, err := s.comments2CommentAPIs(comments, uid, admin, superObj.GetObjID())
	if err != nil {
		return nil, err
	}
	return commentAPIs, nil
}

// DeleteComment 删除评论
func (s *ReactionService) DeleteComment(id uint, uid uint, admin bool) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var comment database.Comment
		if err := tx.Limit(1).Find(&comment, "id = ?", id).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if comment.ID == 0 {
			return errors.New("评论不存在Orz")
		}

		// 不是自己的评论且不是管理员
		if comment.Uid != uid && !admin {
			return errors.New("无权删除Orz")
		}

		// 评论数-1
		obj, err := types.NewObj(comment.Obj)
		if err != nil {
			return err
		}
		if _, err := obj.CommentHandler(tx, comment, -1, uid); err != nil {
			return errors.New("更新评论数失败Orz")
		}

		if err := tx.Delete(&comment).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		return nil
	})
}

// Stay 停留
func (s *ReactionService) Stay(obj types.Obj, uid uint) error {
	if obj.GetObjType() == "poster" {
		return gorse.InsertFeedback(client.Feedback{
			FeedbackType: "stay",
			UserId:       fmt.Sprintf("%d", uid),
			ItemId:       fmt.Sprintf("%d", obj.GetID()),
			Timestamp:    time.Now().String(),
		})
	}
	return errors.New("不能停留Orz")
}
