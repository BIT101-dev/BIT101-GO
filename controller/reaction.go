/*
 * @Author: flwfdd
 * @Date: 2023-03-21 23:16:18
 * @LastEditTime: 2023-03-23 13:33:23
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"errors"
	"fmt"

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
		like_num, err = CommentOnLike(obj_id, delta)
	case "course":
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}
	commit()
	c.JSON(200, gin.H{"like": !like.DeletedAt.Valid, "like_num": like_num})
}

func CheckLike(obj string, uid uint) bool {
	var like database.Like
	database.DB.Where("uid = ?", uid).Where("obj = ?", obj).Limit(1).Find(&like)
	return like.ID != 0
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

func GetCommentList(obj string, order string, page uint, uid uint) []ReactionCommentAPI {
	var list = make([]ReactionCommentAPI, 0)
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
	for _, db_comment := range db_list {
		list = append(list, CleanComment(db_comment, uid))
	}
	return list
}

// 将数据库格式评论转化为返回格式
func CleanComment(old_comment database.Comment, uid uint) ReactionCommentAPI {
	comment_obj := "comment" + fmt.Sprint(old_comment.ID)
	var user UserAPI
	if old_comment.Anonymous {
		user = GetUserAPI(-1)
	} else {
		user = GetUserAPI(int(old_comment.Uid))
	}
	comment := ReactionCommentAPI{
		Comment:   old_comment,
		Like:      CheckLike(comment_obj, uid),
		Own:       old_comment.Uid == uid,
		ReplyUser: GetUserAPI(old_comment.ReplyUid),
		User:      user,
	}

	if comment.CommentNum > 0 {
		comment.Sub = GetCommentList(comment_obj, "like", 0, uid)
		sz := int(config.Config.CommentPreviewSize)
		if len(comment.Sub) > sz {
			comment.Sub = comment.Sub[:sz]
		}
	} else {
		comment.Sub = make([]ReactionCommentAPI, 0)
	}

	return comment
}

// 评论请求结构
type ReactionCommentQuery struct {
	Obj       string `json:"obj" binding:"required"`    // 操作对象
	Text      string `json:"text" binding:"required"`   // 评论内容
	Anonymous bool   `json:"anonymous" default:"false"` // 是否匿名
	ReplyUid  int    `json:"reply_uid" default:"0"`     //回复用户id
	Rate      uint   `json:"rate" default:"0"`          //评分
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
		_, err = CommentOnComment(obj_id, 1)
	case "course":
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
		Rate:      query.Rate,
		Uid:       c.GetUint("uid_uint"),
	}
	database.DB.Create(&comment)

	c.JSON(200, ReactionCommentResponse{
		ReactionCommentAPI: CleanComment(comment, comment.Uid),
		Msg:                "评论成功OvO",
	})
}

// 获取评论列表请求结构
type ReactionCommentListQuery struct {
	Obj   string `form:"obj" binding:"required"` // 操作对象
	Order string `form:"order" default:"old"`    // 排序方式
	Page  uint   `form:"page" default:"0"`       // 页码
}

// 获取评论列表
func ReactionCommentList(c *gin.Context) {
	var query ReactionCommentListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	c.JSON(200, GetCommentList(query.Obj, query.Order, query.Page, c.GetUint("uid_uint")))
}

// 点赞评论
func CommentOnLike(id string, delta int) (uint, error) {
	var comment database.Comment
	database.DB.Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.LikeNum = uint(int(comment.LikeNum) + delta)
	database.DB.Save(&comment)
	return comment.LikeNum, nil
}

// 评论评论
func CommentOnComment(id string, delta int) (uint, error) {
	var comment database.Comment
	database.DB.Limit(1).Find(&comment, "id = ?", id)
	if comment.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	comment.CommentNum = uint(int(comment.CommentNum) + delta)
	database.DB.Save(&comment)
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
		_, err = CommentOnComment(obj_id, -1)
	case "course":
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}

	database.DB.Delete(&comment)
	c.JSON(200, gin.H{"msg": "删除成功OvO"})
}
