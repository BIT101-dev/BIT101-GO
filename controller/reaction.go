/*
 * @Author: flwfdd
 * @Date: 2023-03-21 23:16:18
 * @LastEditTime: 2025-02-08 17:53:37
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/cache"
	"BIT101-GO/util/config"
	"BIT101-GO/util/gorse"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zhenghaoz/gorse/client"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 获取对象的类型和对象的ID
func getTypeID(obj string) (string, string) {
	if obj[:5] == "paper" {
		return "paper", obj[5:]
	}
	if obj[:6] == "poster" {
		return "poster", obj[6:]
	}
	if obj[:6] == "course" {
		return "course", obj[6:]
	}
	if obj[:7] == "comment" {
		return "comment", obj[7:]
	}
	return "", ""
}

// 点赞请求结构
type ReactionLikeQuery struct {
	Obj string `json:"obj" binding:"required"` // 操作对象
}

// 点赞
func ReactionLike(c *gin.Context) {
	var query ReactionLikeQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	obj_type, obj_id := getTypeID(query.Obj)
	if obj_type == "" {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	// 限流
	set, err := cache.RDB.SetNX(cache.Context, "like_uid:"+c.GetString("uid")+"_"+query.Obj, "1", time.Second).Result()
	if err != nil || !set {
		c.JSON(500, gin.H{"msg": "操作过于频繁Orz"})
		return
	}

	database.DB.Transaction(func(tx *gorm.DB) error {
		delta := 0
		var like database.Like
		tx.Unscoped().Where("uid = ?", c.GetString("uid")).Where("obj = ?", query.Obj).Limit(1).Find(&like)
		if like.ID == 0 { //新建
			like = database.Like{
				Obj: query.Obj,
				Uid: c.GetUint("uid_uint"),
			}
			tx.Create(&like)
			delta = 1
		} else if like.DeletedAt.Valid { //删除过 取消删除
			like.DeletedAt = gorm.DeletedAt{}
			tx.Unscoped().Save(like)
			delta = 1

		} else { //取消点赞
			tx.Delete(&like)
			delta = -1
		}

		var like_num uint
		switch obj_type {
		case "paper":
			like_num, err = PaperOnLike(tx, obj_id, delta)
		case "comment":
			like_num, err = CommentOnLike(tx, obj_id, delta, c.GetUint("uid_uint"))
		case "course":
			like_num, err = CourseOnLike(tx, obj_id, delta)
		case "poster":
			like_num, err = PosterOnLike(tx, obj_id, delta, c.GetUint("uid_uint"))
		}
		if err != nil {
			c.JSON(500, gin.H{"msg": "无效对象Orz"})
			return err
		}

		c.JSON(200, gin.H{"like": !like.DeletedAt.Valid, "like_num": like_num})
		return nil
	})
}

// 检查是否点赞
func CheckLike(obj string, uid uint) bool {
	var like database.Like
	database.DB.Where("uid = ?", uid).Where("obj = ?", obj).Limit(1).Find(&like)
	return like.ID != 0
}

// 批量检查是否点赞
func CheckLikeMap(obj_map map[string]bool, uid uint) map[string]bool {
	obj_list := make([]string, 0, len(obj_map))
	for obj := range obj_map {
		obj_list = append(obj_list, obj)
		obj_map[obj] = false
	}
	var like_list []database.Like
	database.DB.Where("uid = ?", uid).Where("obj IN ?", obj_list).Find(&like_list)
	for _, like := range like_list {
		obj_map[like.Obj] = true
	}
	return obj_map
}

// 评论返回结构
type ReactionCommentAPI struct {
	database.Comment
	Like      bool                 `json:"like"`       // 是否点赞
	Own       bool                 `json:"own"`        // 是否是自己的评论
	ReplyUser UserAPI              `json:"reply_user"` // 回复的用户
	Sub       []ReactionCommentAPI `json:"sub"`        // 子评论
	User      UserAPI              `json:"user"`       // 评论用户
	Images    []ImageAPI           `json:"images"`     // 图片
}

// 获取评论列表
func GetCommentList(obj string, order string, page uint, uid uint, admin bool, super_obj string) []ReactionCommentAPI {
	var db_list []database.Comment
	q := database.DB.Model(&database.Comment{}).Where("obj = ?", obj)
	// 排序
	if order == "like" {
		q = q.Order("like_num DESC")
	} else if order == "old" { //发布时间早的在前
		q = q.Order("created_at")
	} else if order == "new" { //发布时间晚的在前
		q = q.Order("created_at DESC")
	} else { //默认 状态新的在前
		q = q.Order("updated_at DESC")
	}
	// 分页
	page_size := config.Config.CommentPageSize
	q = q.Offset(int(page * page_size)).Limit(int(page_size))
	q.Find(&db_list)
	return CleanCommentList(db_list, uid, admin, super_obj)
}

// GetAnonymousName 根据obj和id，hash出独特的匿名序号
func GetAnonymousName(obj string, uid uint) string {
	// 使用hash算法生成匿名序号
	hasher := md5.New()
	hasher.Write([]byte(obj + config.Config.Key + fmt.Sprint(uid)))
	hashBytes := hasher.Sum(nil)
	return "匿名者·" + hex.EncodeToString(hashBytes)[:6]
}

// 将数据库格式评论转化为返回格式
func CleanComment(old_comment database.Comment, uid uint, admin bool, super_obj string) ReactionCommentAPI {
	return CleanCommentList([]database.Comment{old_comment}, uid, admin, super_obj)[0]
}

// 批量将数据库格式评论转化为返回格式
func CleanCommentList(old_comments []database.Comment, uid uint, admin bool, super_obj string) []ReactionCommentAPI {
	comments := make([]ReactionCommentAPI, 0)

	// 查询用户和点赞情况
	uid_map := make(map[int]bool)
	like_map := make(map[string]bool)
	comment_obj_list := make([]string, 0)
	sub_comment_map := make(map[string][]ReactionCommentAPI)
	for _, old_comment := range old_comments {
		// 匿名用户
		if old_comment.Anonymous {
			uid_map[-1] = true
		} else {
			uid_map[int(old_comment.Uid)] = true
		}
		// 回复用户
		if int(old_comment.ReplyUid) < 0 {
			uid_map[-1] = true
			uid_map[-int(old_comment.ReplyUid)] = true
		} else {
			uid_map[int(old_comment.ReplyUid)] = true
		}
		like_map["comment"+fmt.Sprint(old_comment.ID)] = true
		if old_comment.CommentNum > 0 {
			comment_obj_list = append(comment_obj_list, "comment"+fmt.Sprint(old_comment.ID))
		}
		sub_comment_map["comment"+fmt.Sprint(old_comment.ID)] = make([]ReactionCommentAPI, 0)
	}

	// 查询子评论
	var sub_comment_list []database.Comment
	database.DB.Raw(`SELECT * FROM (SELECT *,ROW_NUMBER() OVER (PARTITION BY "obj" ORDER BY "like_num" DESC) AS rn FROM comments WHERE "deleted_at" IS NULL AND obj IN ?) t WHERE rn<=?`, comment_obj_list, config.Config.CommentPreviewSize).Scan(&sub_comment_list)
	for _, sub_comment := range sub_comment_list {
		// 匿名用户
		if sub_comment.Anonymous {
			uid_map[-1] = true
		} else {
			uid_map[int(sub_comment.Uid)] = true
		}
		// 回复用户
		if int(sub_comment.ReplyUid) < 0 {
			uid_map[-1] = true
			uid_map[-int(sub_comment.ReplyUid)] = true
		} else {
			uid_map[int(sub_comment.ReplyUid)] = true
		}
		like_map["comment"+fmt.Sprint(sub_comment.ID)] = true
	}

	// 批量获取用户和点赞
	users := GetUserAPIMap(uid_map)
	likes := CheckLikeMap(like_map, uid)

	// 组装子评论
	for _, sub_comment := range sub_comment_list {
		var user UserAPI
		if sub_comment.Anonymous {
			user = users[-1]
			user.Nickname = GetAnonymousName(super_obj, sub_comment.Uid)
		} else {
			user = users[int(sub_comment.Uid)]
		}
		var reply_user UserAPI
		if int(sub_comment.ReplyUid) < 0 {
			reply_user = users[-1]
			reply_user.Nickname = GetAnonymousName(super_obj, uint(-int(sub_comment.ReplyUid)))
			sub_comment.ReplyUid = -1
		} else if int(sub_comment.ReplyUid) > 0 {
			reply_user = users[int(sub_comment.ReplyUid)]
		}
		sub_comment_map[sub_comment.Obj] = append(sub_comment_map[sub_comment.Obj], ReactionCommentAPI{
			Comment:   sub_comment,
			Like:      likes["comment"+fmt.Sprint(sub_comment.ID)],
			Own:       sub_comment.Uid == uid || admin,
			ReplyUser: reply_user,
			User:      user,
			Sub:       make([]ReactionCommentAPI, 0),
			Images:    GetImageAPIArr(spilt(sub_comment.Images)),
		})
	}

	// 组装评论
	for _, old_comment := range old_comments {
		comment_obj := "comment" + fmt.Sprint(old_comment.ID)
		var user UserAPI
		if old_comment.Anonymous {
			user = users[-1]
			user.Nickname = GetAnonymousName(super_obj, old_comment.Uid)
		} else {
			user = users[int(old_comment.Uid)]
		}
		var reply_user UserAPI
		if int(old_comment.ReplyUid) < 0 {
			reply_user = users[-1]
			reply_user.Nickname = GetAnonymousName(super_obj, uint(-int(old_comment.ReplyUid)))
			old_comment.ReplyUid = -1
		} else if int(old_comment.ReplyUid) > 0 {
			reply_user = users[int(old_comment.ReplyUid)]
		}
		comment := ReactionCommentAPI{
			Comment:   old_comment,
			Like:      likes[comment_obj],
			Own:       old_comment.Uid == uid || admin,
			ReplyUser: reply_user,
			User:      user,
			Images:    GetImageAPIArr(spilt(old_comment.Images)),
		}

		comment.Sub = sub_comment_map[comment_obj]
		comments = append(comments, comment)
	}

	return comments
}

// 评论请求结构
type ReactionCommentQuery struct {
	Obj       string   `json:"obj" binding:"required"` // 操作对象
	Text      string   `json:"text"`                   // 评论内容
	Anonymous bool     `json:"anonymous"`              // 是否匿名
	ReplyUid  int      `json:"reply_uid"`              //回复用户id
	ReplyObj  string   `json:"reply_obj"`              //回复对象
	Rate      uint     `json:"rate"`                   //评分
	ImageMids []string `json:"image_mids"`             //图片
}

// 评论
func ReactionComment(c *gin.Context) {
	// 限流
	set, err := cache.RDB.SetNX(cache.Context, "comment_uid:"+c.GetString("uid"), "1", time.Second).Result()
	if err != nil || !set {
		c.JSON(500, gin.H{"msg": "操作过于频繁Orz"})
		return
	}

	// 参数检查
	var query ReactionCommentQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if query.Text == "" && len(query.ImageMids) == 0 {
		c.JSON(500, gin.H{"msg": "内容不能为空Orz"})
		return
	}
	if !CheckImage(query.ImageMids) {
		c.JSON(500, gin.H{"msg": "存在未上传成功的图片Orz"})
		return
	}
	obj_type, obj_id := getTypeID(query.Obj)
	if obj_type == "" {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	database.DB.Transaction(func(tx *gorm.DB) error {
		// 评价最多一次
		var comment database.Comment
		if query.Rate != 0 {
			tx.Limit(1).Where("uid = ?", c.GetString("uid")).Where("obj = ?", query.Obj).Find(&comment)
			if comment.ID != 0 {
				c.JSON(500, gin.H{"msg": "不能重复评价Orz"})
				return errors.New("不能重复评价Orz")
			}
			time.Sleep(time.Second)
		}

		// 回复匿名用户
		reply_uid := query.ReplyUid
		if reply_uid == -1 {
			var reply_comment database.Comment
			tx.Limit(1).Find(&reply_comment, "id = ?", strings.TrimPrefix(query.ReplyObj, "comment"))
			if reply_comment.ID != 0 {
				reply_uid = -int(reply_comment.Uid)
			} else {
				c.JSON(500, gin.H{"msg": "获取回复用户失败Orz"})
				return errors.New("获取回复用户失败Orz")
			}
		}

		// 评论
		comment = database.Comment{
			Obj:       query.Obj,
			Text:      query.Text,
			Anonymous: query.Anonymous,
			ReplyUid:  reply_uid,
			ReplyObj:  query.ReplyObj,
			Rate:      query.Rate,
			Uid:       c.GetUint("uid_uint"),
			Images:    strings.Join(query.ImageMids, " "),
		}
		tx.Create(&comment)

		// 评论数+1
		var err error
		switch obj_type {
		case "paper":
			_, err = PaperOnComment(tx, obj_id, 1)
		case "comment":
			_, err = CommentOnComment(tx, obj_id, 1, c.GetUint("uid_uint"), query.Anonymous, query.ReplyObj, query.Text)
		case "course":
			_, err = CourseOnComment(tx, obj_id, 1, int(query.Rate))
		case "poster":
			_, err = PosterOnComment(tx, obj_id, 1, c.GetUint("uid_uint"), query.Anonymous, query.Text)
		}
		if err != nil {
			c.JSON(500, gin.H{"msg": "无效对象Orz"})
			return err
		}

		c.JSON(200, CleanComment(comment, comment.Uid, c.GetBool("admin"), query.Obj))
		return nil
	})
}

// 获取评论列表请求结构
type ReactionCommentListQuery struct {
	Obj   string `form:"obj" binding:"required"` // 操作对象
	Order string `form:"order"`                  // 排序方式
	Page  uint   `form:"page"`                   // 页码
}

// 获取评论列表
func ReactionCommentList(c *gin.Context) {
	var query ReactionCommentListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// super_obj用于匿名用户名哈希
	super_obj := query.Obj
	if strings.HasPrefix(query.Obj, "comment") {
		var comment database.Comment
		database.DB.Limit(1).Find(&comment, "id = ?", strings.TrimPrefix(query.Obj, "comment"))
		if comment.ID == 0 {
			c.JSON(500, gin.H{"msg": "评论不存在Orz"})
			return
		}
		super_obj = comment.Obj
	}
	c.JSON(200, GetCommentList(query.Obj, query.Order, query.Page, c.GetUint("uid_uint"), c.GetBool("admin") || c.GetBool("super"), super_obj))
}

// 点赞评论
func CommentOnLike(tx *gorm.DB, id string, delta int, from_uid uint) (uint, error) {
	var comment database.Comment
	tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.LikeNum = uint(int(comment.LikeNum) + delta)
	if err := tx.Save(&comment).Error; err != nil {
		return 0, err
	}

	// 通知
	if delta == 1 && from_uid != comment.Uid {
		go func() {
			// 获取顶级对象 处理子评论问题
			link_obj := comment.Obj
			for strings.HasPrefix(link_obj, "comment") {
				var parent_comment database.Comment
				if err := database.DB.Limit(1).Find(&parent_comment, "id = ?", strings.TrimPrefix(comment.Obj, "comment")).Error; err != nil || parent_comment.ID == 0 {
					return
				}
				link_obj = parent_comment.Obj
			}

			MessageSend(int(from_uid), comment.Uid, fmt.Sprintf("comment%v", comment.ID), "like", link_obj, comment.Text)
		}()
	}

	return comment.LikeNum, nil
}

// 评论评论
func CommentOnComment(tx *gorm.DB, id string, delta int, from_uid uint, from_anonymous bool, reply_obj string, content string) (uint, error) {
	var comment database.Comment
	tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.CommentNum = uint(int(comment.CommentNum) + delta)
	if err := tx.Save(&comment).Error; err != nil {
		return 0, err
	}

	// 通知
	if delta == 1 {
		go func() {
			// 确定通知对象
			to_uid := comment.Uid
			if strings.HasPrefix(reply_obj, "comment") {
				var reply_comment database.Comment
				if err := database.DB.Limit(1).Find(&reply_comment, "id = ?", strings.TrimPrefix(reply_obj, "comment")).Error; err != nil || reply_comment.ID == 0 {
					return
				}
				to_uid = reply_comment.Uid
			}
			if to_uid == from_uid {
				return
			}

			// 获取顶级对象 处理子评论问题
			link_obj := comment.Obj
			for strings.HasPrefix(link_obj, "comment") {
				var parent_comment database.Comment
				if err := database.DB.Limit(1).Find(&parent_comment, "id = ?", strings.TrimPrefix(comment.Obj, "comment")).Error; err != nil || parent_comment.ID == 0 {
					return
				}
				link_obj = parent_comment.Obj
			}

			from_uid_ := int(from_uid)
			if from_anonymous {
				from_uid_ = -1
			}
			MessageSend(from_uid_, to_uid, fmt.Sprintf("comment%v", comment.ID), "comment", link_obj, content)
		}()
	}

	return comment.CommentNum, nil
}

// 删除评论
func ReactionCommentDelete(c *gin.Context) {
	id := c.Param("id")

	database.DB.Transaction(func(tx *gorm.DB) error {
		var comment database.Comment
		tx.Limit(1).Find(&comment, "id = ?", id)
		if comment.ID == 0 {
			c.JSON(500, gin.H{"msg": "评论不存在Orz"})
			return errors.New("评论不存在Orz")
		}

		if comment.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") && !c.GetBool("super") {
			c.JSON(500, gin.H{"msg": "无权删除Orz"})
			return errors.New("无权删除Orz")
		}

		// 评论数-1
		var err error
		obj_type, obj_id := getTypeID(comment.Obj)
		switch obj_type {
		case "paper":
			_, err = PaperOnComment(tx, obj_id, -1)
		case "comment":
			_, err = CommentOnComment(tx, obj_id, -1, 0, true, "", "")
		case "course":
			_, err = CourseOnComment(tx, obj_id, -1, -int(comment.Rate))
		case "poster":
			_, err = PosterOnComment(tx, obj_id, -1, c.GetUint("uid_uint"), true, "")
		}
		if err != nil {
			c.JSON(500, gin.H{"msg": "无效对象Orz"})
			return err
		}

		if err := tx.Delete(&comment).Error; err != nil {
			c.JSON(500, gin.H{"msg": "删除失败Orz"})
			return err
		}
		c.JSON(200, gin.H{"msg": "删除成功OvO"})
		return nil
	})
}

// ReactionStayQuery 停留请求结构
type ReactionStayQuery struct {
	Obj  string `json:"obj" binding:"required"`  // 操作对象
	Time int    `json:"time" binding:"required"` // 停留时间
}

// ReactionStay 停留
func ReactionStay(c *gin.Context) {
	var query ReactionStayQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	obj_type, obj_id := getTypeID(query.Obj)
	if obj_type == "" {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}
	if obj_type == "poster" {
		go func() {
			gorse.InsertFeedback(client.Feedback{
				FeedbackType: "stay",
				UserId:       c.GetString("uid"),
				ItemId:       obj_id,
				Timestamp:    time.Now().String(),
			})
		}()
	}
	c.JSON(200, gin.H{})
}
