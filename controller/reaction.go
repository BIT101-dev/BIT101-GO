/*
 * @Author: flwfdd
 * @Date: 2023-03-21 23:16:18
 * @LastEditTime: 2023-05-17 16:59:19
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取对象的类型和对象的ID
func getTypeID(obj string) (string, string) {
	if obj[:5] == "paper" {
		return "paper", obj[5:]
	}
	if obj[:7] == "comment" {
		return "comment", obj[7:]
	}
	if obj[:6] == "course" {
		return "course", obj[6:]
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

	delta := 0
	var like database.Like
	var commit func()
	database.DB.Unscoped().Where("uid = ?", c.GetString("uid")).Where("obj = ?", query.Obj).Limit(1).Find(&like)
	if like.ID == 0 { //新建
		like = database.Like{
			Obj: query.Obj,
			Uid: c.GetUint("uid_uint"),
		}
		commit = func() { database.DB.Create(&like) }
		delta = 1
	} else if like.DeletedAt.Valid { //删除过 取消删除
		like.DeletedAt = gorm.DeletedAt{}
		commit = func() { database.DB.Unscoped().Save(like) }
		delta = 1

	} else { //取消点赞
		commit = func() { database.DB.Delete(&like) }
		delta = -1
	}

	var like_num uint
	var err error
	switch obj_type {
	case "paper":
		like_num, err = PaperOnLike(obj_id, delta)
	case "comment":
		like_num, err = CommentOnLike(obj_id, delta, c.GetUint("uid_uint"))
	case "course":
		like_num, err = CourseOnLike(obj_id, delta)
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}
	commit()
	c.JSON(200, gin.H{"like": !like.DeletedAt.Valid, "like_num": like_num})
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
}

// 获取评论列表
func GetCommentList(obj string, order string, page uint, uid uint, admin bool) []ReactionCommentAPI {
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
	return CleanCommentList(db_list, uid, admin)
}

// 将数据库格式评论转化为返回格式
func CleanComment(old_comment database.Comment, uid uint, admin bool) ReactionCommentAPI {
	return CleanCommentList([]database.Comment{old_comment}, uid, admin)[0]
}

// 批量将数据库格式评论转化为返回格式
func CleanCommentList(old_comments []database.Comment, uid uint, admin bool) []ReactionCommentAPI {
	comments := make([]ReactionCommentAPI, 0)

	// 查询用户和点赞情况
	uid_map := make(map[int]bool)
	like_map := make(map[string]bool)
	comment_obj_list := make([]string, 0)
	sub_comment_map := make(map[string][]ReactionCommentAPI)
	for _, old_comment := range old_comments {
		if old_comment.Anonymous {
			uid_map[-1] = true
		} else {
			uid_map[int(old_comment.Uid)] = true
		}
		uid_map[int(old_comment.ReplyUid)] = true
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
		if sub_comment.Anonymous {
			uid_map[-1] = true
		} else {
			uid_map[int(sub_comment.Uid)] = true
		}
		uid_map[int(sub_comment.ReplyUid)] = true
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
		} else {
			user = users[int(sub_comment.Uid)]
		}
		sub_comment_map[sub_comment.Obj] = append(sub_comment_map[sub_comment.Obj], ReactionCommentAPI{
			Comment:   sub_comment,
			Like:      likes["comment"+fmt.Sprint(sub_comment.ID)],
			Own:       sub_comment.Uid == uid || admin,
			ReplyUser: users[int(sub_comment.ReplyUid)],
			User:      user,
			Sub:       make([]ReactionCommentAPI, 0),
		})
	}

	// 组装评论
	for _, old_comment := range old_comments {
		comment_obj := "comment" + fmt.Sprint(old_comment.ID)
		var user UserAPI
		if old_comment.Anonymous {
			user = users[-1]
		} else {
			user = users[int(old_comment.Uid)]
		}
		comment := ReactionCommentAPI{
			Comment:   old_comment,
			Like:      likes[comment_obj],
			Own:       old_comment.Uid == uid || admin,
			ReplyUser: users[int(old_comment.ReplyUid)],
			User:      user,
		}

		comment.Sub = sub_comment_map[comment_obj]
		comments = append(comments, comment)
	}

	return comments
}

// 评论请求结构
type ReactionCommentQuery struct {
	Obj       string `json:"obj" binding:"required"`  // 操作对象
	Text      string `json:"text" binding:"required"` // 评论内容
	Anonymous bool   `json:"anonymous"`               // 是否匿名
	ReplyUid  int    `json:"reply_uid"`               //回复用户id
	ReplyObj  string `json:"reply_obj"`               //回复对象
	Rate      uint   `json:"rate"`                    //评分
}

// 评论返回结构
type ReactionCommentResponse struct {
	ReactionCommentAPI
	Msg string `json:"msg"`
}

// 评论
func ReactionComment(c *gin.Context) {
	var query ReactionCommentQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	obj_type, obj_id := getTypeID(query.Obj)
	if obj_type == "" {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	var comment database.Comment
	if query.Rate != 0 {
		database.DB.Limit(1).Where("uid = ?", c.GetString("uid")).Where("obj = ?", query.Obj).Find(&comment)
		if comment.ID != 0 {
			c.JSON(500, gin.H{"msg": "不能重复评价Orz"})
			return
		}
	}

	// 评论数+1
	var err error
	switch obj_type {
	case "paper":
		_, err = PaperOnComment(obj_id, 1)
	case "comment":
		_, err = CommentOnComment(obj_id, 1, c.GetUint("uid_uint"), query.Anonymous, query.ReplyObj, query.Text)
	case "course":
		_, err = CourseOnComment(obj_id, 1, int(query.Rate))
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	// 评论
	comment = database.Comment{
		Obj:       query.Obj,
		Text:      query.Text,
		Anonymous: query.Anonymous,
		ReplyUid:  query.ReplyUid,
		ReplyObj:  query.ReplyObj,
		Rate:      query.Rate,
		Uid:       c.GetUint("uid_uint"),
	}
	database.DB.Create(&comment)

	c.JSON(200, ReactionCommentResponse{
		ReactionCommentAPI: CleanComment(comment, comment.Uid, c.GetBool("admin")),
		Msg:                "评论成功OvO",
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

	c.JSON(200, GetCommentList(query.Obj, query.Order, query.Page, c.GetUint("uid_uint"), c.GetBool("admin")))
}

// 点赞评论
func CommentOnLike(id string, delta int, from_uid uint) (uint, error) {
	var comment database.Comment
	database.DB.Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.LikeNum = uint(int(comment.LikeNum) + delta)
	if err := database.DB.Save(&comment).Error; err != nil {
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

			MessageSend(int(from_uid), comment.Uid, "like", fmt.Sprintf("comment%v", comment.ID), link_obj, comment.Text)
		}()
	}

	return comment.LikeNum, nil
}

// 评论评论
func CommentOnComment(id string, delta int, from_uid uint, from_anonymous bool, reply_obj string, content string) (uint, error) {
	var comment database.Comment
	database.DB.Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.CommentNum = uint(int(comment.CommentNum) + delta)
	if err := database.DB.Save(&comment).Error; err != nil {
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
			MessageSend(from_uid_, to_uid, "comment", fmt.Sprintf("comment%v", comment.ID), link_obj, content)
		}()
	}

	return comment.CommentNum, nil
}

// 删除评论
func ReactionCommentDelete(c *gin.Context) {
	id := c.Param("id")

	var comment database.Comment
	database.DB.Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		c.JSON(500, gin.H{"msg": "评论不存在Orz"})
		return
	}

	if comment.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "无权删除Orz"})
		return
	}

	// 评论数-1
	var err error
	obj_type, obj_id := getTypeID(comment.Obj)
	switch obj_type {
	case "paper":
		_, err = PaperOnComment(obj_id, -1)
	case "comment":
		_, err = CommentOnComment(obj_id, -1, 0, true, "", "")
	case "course":
		_, err = CourseOnComment(obj_id, -1, -int(comment.Rate))
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(500, gin.H{"msg": "删除失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "删除成功OvO"})
}
