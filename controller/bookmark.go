package controller

import (
	"BIT101-GO/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetBookmark(obj string, uid uint) int {
	// Check if the bookmark exists
	var bookmark database.Bookmark
	database.DB.Where("obj = ?", obj).Where("uid = ?", uid).Limit(1).Find(&bookmark)

	if bookmark.ID == 0 {
		return -1
	}

	return bookmark.Type
}

type BookmarksQuery struct {
	Uid int `json:"uid" binding:"required"` // 操作对象
}

type BookmarksResponseItem struct {
	ID          uint   `json:"id"`
	Obj         string `json:"obj"`
	Own         bool   `json:"own"`
	Type        int    `json:"type"`
	Content     string `json:"content"`
	ActivatedAt string `json:"activated_at"`
}

func GetUserBookmarks(c *gin.Context) {
	var query BookmarksQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}

	bookmarks, err := GetBookmarks(uint(query.Uid))
	if err != nil {
		c.JSON(500, gin.H{"msg": "获取失败Orz"})
		return
	}

	resp := make([]BookmarksResponseItem, 0)
	for _, bookmark := range bookmarks {
		resp = append(resp, BookmarksResponseItem{
			ID:          bookmark.ID,
			Obj:         bookmark.Obj,
			Own:         bookmark.Type == 9,
			Type:        bookmark.Type,
			Content:     bookmark.Content,
			ActivatedAt: bookmark.ActivatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(200, resp)
}

type DeleteBookmarkQuery struct {
	Obj string `json:"obj" binding:"required"` // 操作对象
}

type InsertBookmarkQuery struct {
	Obj     string `json:"obj" binding:"required"`     // 操作对象
	Type    int    `json:"type" binding:"required"`    // 类型
	Content string `json:"content" binding:"required"` // 内容
}

func SetUserBookmark(c *gin.Context) {
	var query InsertBookmarkQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}
	if err := SetBookmark(query.Obj, c.GetUint("uid_uint"), query.Type, query.Content); err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "操作成功OvO"})
}

func DeleteUserBookmark(c *gin.Context) {
	var query DeleteBookmarkQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}
	if err := DeleteBookmark(query.Obj, c.GetUint("uid_uint")); err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "操作成功OvO"})
}

func GetBookmarks(uid uint) ([]database.Bookmark, error) {
	var bookmarks []database.Bookmark
	if err := database.DB.Where("uid = ?", uid).Find(&bookmarks).Error; err != nil {
		return nil, err
	}
	return bookmarks, nil
}

func GetRelatedBookmarks(obj string) ([]database.Bookmark, error) {
	var bookmarks []database.Bookmark
	if err := database.DB.Where("obj = ?", obj).Find(&bookmarks).Error; err != nil {
		return nil, err
	}
	return bookmarks, nil
}

func SetBookmark(obj string, uid uint, t int, content string) error {
	// Check if the bookmark exists
	var bookmark database.Bookmark
	database.DB.Unscoped().Where("obj = ?", obj).Where("uid = ?", uid).Limit(1).Find(&bookmark)

	bookmark.Obj = obj
	bookmark.Uid = uid
	bookmark.Type = t
	bookmark.Content = content

	var err error
	if bookmark.ID == 0 {
		err = database.DB.Create(&bookmark).Error
	} else if bookmark.DeletedAt.Valid { //删除过 取消删除
		bookmark.DeletedAt = gorm.DeletedAt{}
		err = database.DB.Unscoped().Save(bookmark).Error
	} else {
		err = database.DB.Save(&bookmark).Error
	}

	if err != nil {
		return err
	}

	obj_type, obj_id := GetTypeID(obj)
	switch obj_type {
	case "poster":
		if err := PosterOnBookmark(obj_id, 1); err != nil {
			return err
		}
	}

	return nil
}

func DeleteBookmark(obj string, uid uint) error {
	var bookmark database.Bookmark
	database.DB.Where("obj = ?", obj).Where("uid = ?", uid).Limit(1).Find(&bookmark)
	if err := database.DB.Delete(&bookmark).Error; err != nil {
		return err
	}

	obj_type, obj_id := GetTypeID(obj)
	switch obj_type {
	case "poster":
		if err := PosterOnBookmark(obj_id, -1); err != nil {
			return err
		}
	}

	return nil
}
