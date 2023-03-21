/*
 * @Author: flwfdd
 * @Date: 2023-03-21 23:16:18
 * @LastEditTime: 2023-03-22 00:45:03
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"

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
	case "course":
	}
	if err != nil {
		c.JSON(500, gin.H{"msg": "无效对象Orz"})
		return
	}
	commit()
	c.JSON(200, gin.H{"like": !like.DeletedAt.Valid, "like_num": like_num})
}

func CheckLike(obj string, uid string) bool {
	var like database.Like
	database.DB.Where("uid = ?", uid).Where("obj = ?", obj).Limit(1).Find(&like)
	return like.ID != 0
}
