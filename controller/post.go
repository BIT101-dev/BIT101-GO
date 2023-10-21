/**
* @author:YHCnb
* Package:
* @date:2023/10/16 21:52
* Description:
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/gorse"
	"BIT101-GO/util/search"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhenghaoz/gorse/client"
	"strings"
	"time"
)

// PosterGetResponse 获取帖子返回结构
type PosterGetResponse struct {
	database.Poster
	UserAPI `json:"user"`
	Images  []ImageAPI     `json:"images"`
	Tags    []string       `json:"tags"`
	Claim   database.Claim `json:"claim"`
	Like    bool           `json:"like"`
	Own     bool           `json:"own"`
}

func spilt(str string) []string {
	l := strings.Split(str, " ")
	out := make([]string, 0)
	for i := range l {
		if l[i] != "" {
			out = append(out, l[i])
		}
	}
	return out
}

// PosterGet 获取帖子
func PosterGet(c *gin.Context) {
	var post database.Poster
	if err := database.DB.Limit(1).Find(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	// 帖子不可见也不返回
	if post.ID == 0 || !post.Public {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	own := post.Uid == c.GetUint("uid_uint") || c.GetBool("admin")
	var userAPI UserAPI
	if post.Anonymous {
		userAPI = GetUserAPI(-1)
		post.Uid = 0
	} else {
		userAPI = GetUserAPI(int(post.Uid))
	}
	var res = PosterGetResponse{
		Poster:  post,
		UserAPI: userAPI,
		Images:  GetImageAPIArr(spilt(post.Images)),
		Tags:    spilt(post.Tags),
		Claim:   database.ClaimMap[post.ClaimID],
		Like:    CheckLike(fmt.Sprintf("post%v", post.ID), c.GetUint("uid_uint")),
		Own:     own,
	}
	c.JSON(200, res)
}

// PosterUpdateQuery 发布帖子请求接口
type PosterUpdateQuery struct {
	Title     string   `json:"title" binding:"required"`
	Text      string   `json:"text" binding:"required"`
	ImageMids []string `json:"image_mids"`
	Plugins   string   `json:"plugins"`
	Anonymous bool     `json:"anonymous"`
	Tags      []string `json:"tags"`
	ClaimID   uint     `json:"claim_id"`
	Public    bool     `json:"public"`
}

// PosterSubmit 发布帖子
func PosterSubmit(c *gin.Context) {
	var query PosterUpdateQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(500, gin.H{"msg": "参数错误Orz"})
		return
	}
	if _, ok := database.ClaimMap[query.ClaimID]; !ok {
		c.JSON(500, gin.H{"msg": "申明不存在Orz"})
		return
	}

	var post = database.Poster{
		Title:      query.Title,
		Text:       query.Text,
		Images:     strings.Join(query.ImageMids, " "),
		Uid:        c.GetUint("uid_uint"),
		Anonymous:  query.Anonymous,
		Public:     query.Public,
		LikeNum:    0,
		CommentNum: 0,
		Tags:       strings.Join(query.Tags, " "),
		ClaimID:    query.ClaimID,
		Plugins:    query.Plugins,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot([]string{}, query.Tags)
	if err := search.Update("post", post); err != nil {
		c.JSON(500, gin.H{"msg": "search同步失败Orz"})
		return
	}
	if err := gorse.InsertPost(post); err != nil {
		c.JSON(500, gin.H{"msg": "gorse同步失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "发布成功", "id": post.ID})
}

// PosterPut 修改帖子
func PosterPut(c *gin.Context) {
	var query PosterUpdateQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(500, gin.H{"msg": "参数错误Orz"})
		return
	}
	var post database.Poster
	if err := database.DB.Limit(1).Find(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if post.ID == 0 {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	if post.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有修改权限Orz"})
		return
	}
	if _, ok := database.ClaimMap[query.ClaimID]; !ok {
		c.JSON(500, gin.H{"msg": "申明不存在Orz"})
		return
	}
	post.Title = query.Title
	post.Text = query.Text
	post.Images = strings.Join(query.ImageMids, " ")
	post.Tags = strings.Join(query.Tags, " ")
	post.Anonymous = query.Anonymous
	post.Public = query.Public
	post.ClaimID = query.ClaimID
	post.Plugins = query.Plugins
	post.EditAt = time.Now()
	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot(strings.Split(post.Tags, " "), query.Tags)
	if err := search.Update("post", post); err != nil {
		c.JSON(500, gin.H{"msg": "search同步失败Orz"})
		return
	}
	if err := gorse.UpdatePost(post); err != nil {
		c.JSON(500, gin.H{"msg": "gorse同步失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "编辑成功OvO"})
}

// PosterListQuery 获取帖子列表
type PosterListQuery struct {
	Mode   string `form:"mode"` //recommend | search | follow | hot 默认为recommend
	Page   uint   `form:"page"`
	Search string `form:"search"`
	Order  string `form:"order"` //like | new 默认为new
	Uid    int    `form:"uid"`
}

// PosterListResponseItem 获取帖子列表返回结构
type PosterListResponseItem struct {
	database.Poster
	UserAPI `json:"user"`
	Images  []ImageAPI     `json:"images"`
	Tags    []string       `json:"tags"`
	Claim   database.Claim `json:"claim"`
}

// 构建获取帖子列表返回结构
func buildPostListResponse(posts []database.Poster) []PosterListResponseItem {
	res := make([]PosterListResponseItem, 0)
	for _, post := range posts {
		var userAPI UserAPI
		if post.Anonymous {
			userAPI = GetUserAPI(-1)
			post.Uid = 0
		} else {
			userAPI = GetUserAPI(int(post.Uid))
		}
		var claim database.Claim
		database.DB.Limit(1).Find(&claim, "id = ?", post.ClaimID)

		res = append(res, PosterListResponseItem{
			Poster:  post,
			UserAPI: userAPI,
			Images:  GetImageAPIArr(spilt(post.Images)),
			Tags:    spilt(post.Tags),
			Claim:   claim,
		})
	}
	return res
}

// PostList 获取帖子列表
func PostList(c *gin.Context) {
	var query PosterListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	// recommend/hot模式
	if query.Mode == "" || query.Mode == "recommend" {
		recommend, err := gorse.GetRecommend(c.GetString("uid"), query.Page)
		if err != nil {
			c.JSON(500, gin.H{"msg": "获取推荐失败Orz"})
			return
		}
		for i := range recommend {
			println(recommend[i])
		}
		var posts []database.Poster
		database.DB.Find(&posts, "id IN ?", recommend)
		c.JSON(200, buildPostListResponse(posts))
		return
	} else if query.Mode == "hot" {
		popular, err := gorse.GetPopular(c.GetString("uid"), query.Page)
		if err != nil {
			c.JSON(500, gin.H{"msg": "获取热榜失败Orz"})
			return
		}
		for i := range popular {
			println(popular[i])
		}
		var posts []database.Poster
		database.DB.Find(&posts, "id IN ?", popular)
		c.JSON(200, buildPostListResponse(posts))
		return
	}
	// follow/search模式
	var order []string
	if query.Order == "like" {
		order = append(order, "like_num:desc")
	} else {
		order = append(order, "edit_time:desc")
	}
	var filter []string
	if query.Mode == "follow" {
		var follow_list []database.Follow
		database.DB.Where("uid = ?", c.GetString("uid")).Find(&follow_list)
		uid_filter := "uid IN [ "
		for _, follow := range follow_list {
			uid_filter += fmt.Sprintf("%v,", follow.FollowUid)
		}
		uid_filter = uid_filter[:len(uid_filter)-1] + "]"
		filter = append(filter, uid_filter)
		filter = append(filter, "anonymous = false")
		filter = append(filter, "public = true")
	} else {
		switch query.Uid {
		case 0:
			filter = append(filter, "uid ="+c.GetString("uid"))
		case -1:
			filter = append(filter, "public = true")
		default:
			filter = append(filter, "uid ="+fmt.Sprintf("%v", query.Uid))
			filter = append(filter, "anonymous = false")
			filter = append(filter, "public = true")
		}
	}
	var posts []database.Poster
	err := search.Search(&posts, "post", query.Search, query.Page, config.Config.PostPageSize, order, filter)
	if err != nil {
		c.JSON(500, gin.H{"msg": "搜索失败Orz"})
		println(err.Error())
		return
	}
	c.JSON(200, buildPostListResponse(posts))
}

// PosterDelete 删除帖子
func PosterDelete(c *gin.Context) {
	var post database.Poster
	if err := database.DB.Limit(1).Find(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if post.ID == 0 {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	if post.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有删除权限Orz"})
		return
	}
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot(strings.Split(post.Tags, " "), []string{})
	if err := search.Delete("post", []string{fmt.Sprintf("%v", post.ID)}); err != nil {
		c.JSON(500, gin.H{"msg": "search同步失败Orz"})
		return
	}
	if err := gorse.DeletePost(c.Param("id")); err != nil {
		c.JSON(500, gin.H{"msg": "gorse同步失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "删除成功OvO"})
}

// ClaimList 获取申明列表
func ClaimList(c *gin.Context) {
	var claims []database.Claim
	if err := database.DB.Find(&claims).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, claims)
}

// updateTagHot 更新tag的热度
func updateTagHot(oldTags []string, newTags []string) {
	for _, tag := range oldTags {
		var t database.Tag
		database.DB.Limit(1).Find(&t, "name = ?", tag)
		if t.ID != 0 {
			database.DB.Model(&t).Update("hot", t.Hot-1)
		}
	}
	for _, tag := range newTags {
		var t database.Tag
		database.DB.Limit(1).Find(&t, "name = ?", tag)
		if t.ID == 0 {
			database.DB.Create(&database.Tag{Name: tag, Hot: 1})
		} else {
			database.DB.Model(&t).Update("hot", t.Hot+1)
		}
	}
}

// PosterOnLike 点赞
func PosterOnLike(id string, delta int, uid string) (uint, error) {
	var post database.Poster
	if err := database.DB.Limit(1).Find(&post, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if post.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	post.LikeNum = uint(int(post.LikeNum) + delta)
	if err := database.DB.Save(&post).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}

	var err error
	if delta > 0 {
		err = gorse.InsertFeedback(client.Feedback{
			FeedbackType: "like",
			UserId:       uid,
			ItemId:       id,
			Timestamp:    time.Now().String(),
		})
	} else {
		err = gorse.DeleteFeedback("like", uid, id)
	}
	if err != nil {
		return 0, errors.New("gorse同步失败Orz")
	}
	if err := search.Update("post", post); err != nil {
		return 0, errors.New("search同步失败Orz")
	}
	return post.LikeNum, nil
}

// PosterOnComment 评论
func PosterOnComment(id string, delta int, uid string) (uint, error) {
	var post database.Poster
	if err := database.DB.Limit(1).Find(&post, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if post.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	post.CommentNum = uint(int(post.CommentNum) + delta)
	if err := database.DB.Save(&post).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}

	var err error
	if delta > 0 {
		err = gorse.InsertFeedback(client.Feedback{
			FeedbackType: "comment",
			UserId:       uid,
			ItemId:       id,
			Timestamp:    time.Now().String(),
		})
	} else {
		err = gorse.DeleteFeedback("comment", uid, id)
	}
	if err != nil {
		return 0, errors.New("gorse同步失败Orz")
	}
	if err := search.Update("post", post); err != nil {
		return 0, errors.New("search同步失败Orz")
	}
	return post.CommentNum, nil
}
