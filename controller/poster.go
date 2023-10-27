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
	"strconv"
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

// CheckTags 检查tags
func CheckTags(tags []string) bool {
	for i := range tags {
		if len(tags[i]) > 30 {
			return false
		}
	}
	return true
}

// PosterGet 获取帖子
func PosterGet(c *gin.Context) {
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	// 帖子不可见也不返回
	if poster.ID == 0 || (!poster.Public && poster.Uid != c.GetUint("uid_uint") && !c.GetBool("admin")) {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	own := poster.Uid == c.GetUint("uid_uint") || c.GetBool("admin")
	var userAPI UserAPI
	if poster.Anonymous {
		userAPI = GetUserAPI(-1)
		poster.Uid = 0
	} else {
		userAPI = GetUserAPI(int(poster.Uid))
	}
	var res = PosterGetResponse{
		Poster:  poster,
		UserAPI: userAPI,
		Images:  GetImageAPIArr(spilt(poster.Images)),
		Tags:    spilt(poster.Tags),
		Claim:   database.ClaimMap[poster.ClaimID],
		Like:    CheckLike(fmt.Sprintf("poster%v", poster.ID), c.GetUint("uid_uint")),
		Own:     own,
	}
	c.JSON(200, res)
}

// PosterUpdateQuery 发布帖子请求接口
type PosterUpdateQuery struct {
	Title     string   `json:"title" binding:"required"`
	Text      string   `json:"text"`
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
	if query.Text == "" && len(query.ImageMids) == 0 {
		c.JSON(500, gin.H{"msg": "内容不能为空Orz"})
		return
	}
	if _, ok := database.ClaimMap[query.ClaimID]; !ok {
		c.JSON(500, gin.H{"msg": "申明不存在Orz"})
		return
	}
	if !CheckTags(query.Tags) {
		c.JSON(400, gin.H{"msg": "tags过长Orz"})
		return
	}
	if !CheckImage(query.ImageMids) {
		c.JSON(400, gin.H{"msg": "存在未上传成功的图片Orz"})
		return
	}

	var poster = database.Poster{
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
	if err := database.DB.Create(&poster).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot([]string{}, query.Tags)
	go func() {
		search.Update("poster", poster)
		gorse.InsertPoster(poster)
	}()
	c.JSON(200, gin.H{"msg": "发布成功", "id": poster.ID})
}

// PosterPut 修改帖子
func PosterPut(c *gin.Context) {
	var query PosterUpdateQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(500, gin.H{"msg": "参数错误Orz"})
		return
	}
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if poster.ID == 0 {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	if poster.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有修改权限Orz"})
		return
	}
	if query.Text == "" && len(query.ImageMids) == 0 {
		c.JSON(500, gin.H{"msg": "内容不能为空Orz"})
		return
	}
	if _, ok := database.ClaimMap[query.ClaimID]; !ok {
		c.JSON(500, gin.H{"msg": "申明不存在Orz"})
		return
	}
	if !CheckTags(query.Tags) {
		c.JSON(400, gin.H{"msg": "tags过长Orz"})
		return
	}
	if !CheckImage(query.ImageMids) {
		c.JSON(400, gin.H{"msg": "存在未上传成功的图片Orz"})
		return
	}
	poster.Title = query.Title
	poster.Text = query.Text
	poster.Images = strings.Join(query.ImageMids, " ")
	poster.Tags = strings.Join(query.Tags, " ")
	poster.Anonymous = query.Anonymous
	poster.Public = query.Public
	poster.ClaimID = query.ClaimID
	poster.Plugins = query.Plugins
	poster.EditAt = time.Now()
	if err := database.DB.Save(&poster).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot(strings.Split(poster.Tags, " "), query.Tags)
	go func() {
		search.Update("poster", poster)
		gorse.UpdatePoster(poster)
	}()
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

// 批量获取帖子
func getPostersMap(ids []string) map[string]database.Poster {
	posters := make([]database.Poster, 0)
	database.DB.Find(&posters, "id IN ?", ids)
	res := make(map[string]database.Poster)
	for _, poster := range posters {
		res[strconv.Itoa(int(poster.ID))] = poster
	}
	return res
}

// 构建获取帖子列表返回结构
func buildPostListResponse(posters []database.Poster) []PosterListResponseItem {
	res := make([]PosterListResponseItem, 0)
	for _, poster := range posters {
		var userAPI UserAPI
		if poster.Anonymous {
			userAPI = GetUserAPI(-1)
			poster.Uid = 0
		} else {
			userAPI = GetUserAPI(int(poster.Uid))
		}
		var claim database.Claim
		database.DB.Limit(1).Find(&claim, "id = ?", poster.ClaimID)

		res = append(res, PosterListResponseItem{
			Poster:  poster,
			UserAPI: userAPI,
			Images:  GetImageAPIArr(spilt(poster.Images)),
			Tags:    spilt(poster.Tags),
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
		var posters []database.Poster
		postersMap := getPostersMap(recommend)
		// 按照推荐顺序获取帖子
		for _, item := range recommend {
			if _, ok := postersMap[item]; ok {
				posters = append(posters, postersMap[item])
			}
		}
		c.JSON(200, buildPostListResponse(posters))
		return
	} else if query.Mode == "hot" {
		popular, err := gorse.GetPopular(c.GetString("uid"), query.Page)
		if err != nil {
			c.JSON(500, gin.H{"msg": "获取热榜失败Orz"})
			return
		}
		var posters []database.Poster
		postersMap := getPostersMap(popular)
		// 按照推荐顺序获取帖子
		for _, item := range popular {
			if _, ok := postersMap[item]; ok {
				posters = append(posters, postersMap[item])
			}
		}
		c.JSON(200, buildPostListResponse(posters))
		return
	}
	// follow/search模式
	var order []string
	if query.Order == "like" {
		order = append(order, "like_num:desc")
	} else if query.Order == "new" {
		order = append(order, "edit_time:desc")
	}
	var filter []string
	if query.Mode == "follow" {
		query.Search = ""
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
	var posters []database.Poster
	err := search.Search(&posters, "poster", query.Search, query.Page, config.Config.PostPageSize, order, filter)
	if err != nil {
		c.JSON(500, gin.H{"msg": "搜索失败Orz"})
		println(err.Error())
		return
	}
	c.JSON(200, buildPostListResponse(posters))
}

// PosterDelete 删除帖子
func PosterDelete(c *gin.Context) {
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if poster.ID == 0 {
		c.JSON(500, gin.H{"msg": "帖子不存在Orz"})
		return
	}
	if poster.Uid != c.GetUint("uid_uint") && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有删除权限Orz"})
		return
	}
	if err := database.DB.Delete(&poster).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	updateTagHot(strings.Split(poster.Tags, " "), []string{})
	go func() {
		search.Delete("poster", []string{fmt.Sprintf("%v", poster.ID)})
		gorse.DeletePoster(c.Param("id"))
	}()
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
func PosterOnLike(id string, delta int, from_uid uint) (uint, error) {
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if poster.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	poster.LikeNum = uint(int(poster.LikeNum) + delta)
	if err := database.DB.Save(&poster).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		if delta > 0 {
			gorse.InsertFeedback(client.Feedback{
				FeedbackType: "like",
				UserId:       strconv.Itoa(int(from_uid)),
				ItemId:       id,
				Timestamp:    time.Now().Add(time.Duration(config.Config.Gorse.SBEFB) * time.Second).String(), //一段时间后生效,
			})
			if from_uid != poster.Uid {
				post_obj := fmt.Sprintf("poster%v", poster.ID)
				MessageSend(int(from_uid), poster.Uid, post_obj, "like", post_obj, poster.Title)
			}
		} else {
			gorse.DeleteFeedback("like", strconv.Itoa(int(from_uid)), id)
		}
		search.Update("poster", poster)
	}()
	return poster.LikeNum, nil
}

// PosterOnComment 评论
func PosterOnComment(id string, delta int, from_uid uint, from_anonymous bool, content string) (uint, error) {
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if poster.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	poster.CommentNum = uint(int(poster.CommentNum) + delta)
	if err := database.DB.Save(&poster).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}

	go func() {
		if delta > 0 {
			gorse.InsertFeedback(client.Feedback{
				FeedbackType: "comment",
				UserId:       strconv.Itoa(int(from_uid)),
				ItemId:       id,
				Timestamp:    time.Now().Add(time.Duration(config.Config.Gorse.SBEFB) * time.Second).String(), //一段时间后生效,
			})
			if from_uid != poster.Uid {
				from_uid_ := int(from_uid)
				if from_anonymous {
					from_uid_ = -1
				}
				post_obj := fmt.Sprintf("poster%v", poster.ID)
				MessageSend(from_uid_, poster.Uid, post_obj, "comment", post_obj, content)
			}
		} else {
			gorse.DeleteFeedback("comment", strconv.Itoa(int(from_uid)), id)
		}
		search.Update("poster", poster)
	}()
	return poster.CommentNum, nil
}
