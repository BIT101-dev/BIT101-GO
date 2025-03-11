/*
 * @Author: flwfdd
 * @Date: 2025-03-11 12:20:22
 * @LastEditTime: 2025-03-11 19:14:21
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/gorse"
	"BIT101-GO/pkg/search"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zhenghaoz/gorse/client"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ types.PosterService = (*PosterService)(nil)

type PosterService struct {
	userSvc     types.UserService
	imageSvc    types.ImageService
	reactionSvc types.ReactionService
	messageSvc  types.MessageService
}

func NewPosterService(userSvc types.UserService, imageSvc types.ImageService, reactionSvc types.ReactionService, messageSvc types.MessageService) *PosterService {
	s := PosterService{userSvc: userSvc, imageSvc: imageSvc, reactionSvc: reactionSvc, messageSvc: messageSvc}
	types.RegisterObjHandler(&s)
	return &s
}

/* ObjHandler */

// IsExist 判断帖子是否存在
func (s *PosterService) IsExist(id uint) bool {
	_, err := s.getPoster(id)
	return err == nil
}

// GetObjType 获取帖子类型
func (s *PosterService) GetObjType() string {
	return "poster"
}

// GetUid 获取帖子作者
func (s *PosterService) GetUid(id uint) (uint, error) {
	poster, err := s.getPoster(id)
	if err != nil {
		return 0, err
	}
	return poster.Uid, nil
}

// LikeHandler 点赞帖子
func (s *PosterService) LikeHandler(tx *gorm.DB, id uint, delta int, fromUid uint) (likeNum uint, err error) {
	var poster database.Poster
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&poster, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if poster.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	poster.LikeNum = uint(int(poster.LikeNum) + delta)
	if err := tx.Save(&poster).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		search.Update(s.GetObjType(), poster)
		if delta > 0 {
			gorse.InsertFeedback(client.Feedback{
				FeedbackType: "like",
				UserId:       strconv.Itoa(int(fromUid)),
				ItemId:       strconv.Itoa(int(id)),
				Timestamp:    time.Now().String(),
			})
			if fromUid != poster.Uid {
				objID := fmt.Sprintf("%s%v", s.GetObjType(), id)
				s.messageSvc.Send(int(fromUid), poster.Uid, objID, types.MessageTypeLike, objID, poster.Title)
			}
		} else {
			gorse.DeleteFeedback("like", strconv.Itoa(int(fromUid)), strconv.Itoa(int(id)))
		}
	}()
	return poster.LikeNum, nil
}

// CommentHandler 评论帖子
func (s *PosterService) CommentHandler(tx *gorm.DB, id uint, comment database.Comment, delta int, fromUid uint) (commentNum uint, err error) {
	var poster database.Poster
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&poster, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if poster.ID == 0 {
		return 0, errors.New("帖子不存在Orz")
	}
	poster.CommentNum = uint(int(poster.CommentNum) + delta)
	if err := tx.Save(&poster).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}

	go func() {
		search.Update(s.GetObjType(), poster)
		if delta > 0 {
			gorse.InsertFeedback(client.Feedback{
				FeedbackType: "comment",
				UserId:       strconv.Itoa(int(fromUid)),
				ItemId:       strconv.Itoa(int(id)),
				Timestamp:    time.Now().String(),
			})
			if fromUid != poster.Uid {
				from_uid_ := int(fromUid)
				if comment.Anonymous {
					from_uid_ = -1
				}
				objID := fmt.Sprintf("%s%v", s.GetObjType(), id)
				s.messageSvc.Send(from_uid_, poster.Uid, objID, "comment", objID, comment.Text)
			}
		} else {
			gorse.DeleteFeedback("comment", strconv.Itoa(int(fromUid)), strconv.Itoa(int(id)))
		}
	}()
	return poster.CommentNum, nil
}

// Get 获取帖子
func (s *PosterService) getPoster(id uint) (database.Poster, error) {
	var poster database.Poster
	if err := database.DB.Limit(1).Find(&poster, "id = ?", id).Error; err != nil {
		return poster, errors.New("数据库错误Orz")
	}
	if poster.ID == 0 {
		return poster, errors.New("帖子不存在Orz")
	}
	return poster, nil
}

// Get 获取帖子
func (s *PosterService) Get(id, uid uint, admin bool) (types.PosterInfo, error) {
	poster, err := s.getPoster(id)
	if err != nil {
		return types.PosterInfo{}, err
	}
	own := poster.Uid == uid || admin
	var userAPI types.UserAPI
	if poster.Anonymous {
		userAPI = s.userSvc.GetAnonymousUserAPI()
		userAPI.Nickname = s.userSvc.GetAnonymousName(poster.Uid, fmt.Sprintf("%s%v", s.GetObjType(), id))
		poster.Uid = 0
	} else {
		userAPI, err = s.userSvc.GetUserAPI(int(poster.Uid))
		if err != nil {
			return types.PosterInfo{}, errors.New("获取用户信息失败Orz")
		}
	}
	var posterInfo = types.PosterInfo{
		Poster: poster,
		User:   userAPI,
		Images: s.imageSvc.GetImageAPIList(common.Spilt(poster.Images)),
		Tags:   common.Spilt(poster.Tags),
		Claim:  database.ClaimMap[poster.ClaimID],
		Like:   s.reactionSvc.CheckLike(fmt.Sprintf("%s%v", s.GetObjType(), poster.ID), uid),
		Own:    own,
	}
	return posterInfo, nil
}

// checkTags 检查tags
func checkTags(tags []string) bool {
	if len(tags) > 10 {
		return false
	}
	for i := range tags {
		// 去除空格
		tags[i] = strings.ReplaceAll(tags[i], " ", "")
		if len(tags[i]) > 30 {
			return false
		}
	}
	return true
}

// Create 创建帖子
func (s *PosterService) Create(title, text string, imageMids []string, plugins string, anonymous bool, tags []string, claimID uint, public bool, uid uint, admin bool) (uint, error) {
	if text == "" && len(imageMids) == 0 {
		return 0, errors.New("内容不能为空Orz")
	}
	if _, ok := database.ClaimMap[claimID]; !ok {
		return 0, errors.New("声明不存在Orz")
	}
	if !checkTags(tags) {
		return 0, errors.New("tags过长Orz")
	}
	if !s.imageSvc.CheckMids(imageMids) {
		return 0, errors.New("存在未上传成功的图片Orz")
	}

	var poster = database.Poster{
		Title:      title,
		Text:       text,
		Images:     strings.Join(imageMids, " "),
		Uid:        uid,
		Anonymous:  anonymous,
		Public:     public,
		LikeNum:    0,
		CommentNum: 0,
		Tags:       strings.Join(tags, " "),
		ClaimID:    claimID,
		Plugins:    plugins,
	}
	if err := database.DB.Create(&poster).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		search.Update(s.GetObjType(), poster)
		gorse.InsertPoster(poster)
	}()
	return poster.ID, nil
}

// Edit 修改帖子
func (s *PosterService) Edit(id uint, title, text string, imageMids []string, plugins string, anonymous bool, tags []string, claimID uint, public bool, uid uint, admin bool) error {
	poster, err := s.getPoster(id)
	if err != nil {
		return err
	}
	if poster.Uid != uid && !admin {
		return errors.New("没有修改权限Orz")
	}
	if text == "" && len(imageMids) == 0 {
		return errors.New("内容不能为空Orz")
	}
	if _, ok := database.ClaimMap[claimID]; !ok {
		return errors.New("声明不存在Orz")
	}
	if !checkTags(tags) {
		return errors.New("tags过长Orz")
	}
	if !s.imageSvc.CheckMids(imageMids) {
		return errors.New("存在未上传成功的图片Orz")
	}
	poster.Title = title
	poster.Text = text
	poster.Images = strings.Join(imageMids, " ")
	poster.Tags = strings.Join(tags, " ")
	poster.Anonymous = anonymous
	poster.Public = public
	poster.ClaimID = claimID
	poster.Plugins = plugins
	poster.EditAt = time.Now()
	if err := database.DB.Save(&poster).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	go func() {
		search.Update(s.GetObjType(), poster)
		gorse.UpdatePoster(poster)
	}()
	return nil
}

// 批量获取帖子
func (s *PosterService) getPosters(ids []uint) ([]database.Poster, error) {
	posters := make([]database.Poster, 0)
	if err := database.DB.Find(&posters, "id IN ?", ids).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	var postersMap = make(map[uint]database.Poster)
	for _, poster := range posters {
		postersMap[poster.ID] = poster
	}
	// 按照ids顺序返回
	var res = make([]database.Poster, 0)
	for _, id := range ids {
		if poster, ok := postersMap[id]; ok {
			res = append(res, poster)
		} else {
			return nil, errors.New("帖子不存在Orz")
		}
	}
	return res, nil
}

// 构建获取帖子列表返回结构
func (s *PosterService) posters2PosterAPIs(posters []database.Poster) []types.PosterAPI {
	posterAPIs, err := FillUsers(
		s.userSvc,
		posters,
		func(poster database.Poster) int {
			if poster.Anonymous {
				return -1
			}
			return int(poster.Uid)
		},
		func(poster database.Poster, user types.UserAPI) types.PosterAPI {
			claim := database.ClaimMap[poster.ClaimID]
			poster.Uid = 0
			return types.PosterAPI{
				Poster: poster,
				User:   user,
				Images: s.imageSvc.GetImageAPIList(common.Spilt(poster.Images)),
				Tags:   common.Spilt(poster.Tags),
				Claim:  claim,
			}
		})
	if err != nil {
		return nil
	}
	return posterAPIs
}

// GetList 获取帖子列表
func (s *PosterService) GetList(mode string, page uint, keyword, order string, uid uint, noAnonymous bool, onlyPublic bool) ([]types.PosterAPI, error) {
	// recommend/hot模式 通过推荐系统获取
	if mode == "" || mode == "recommend" {
		ids, err := gorse.GetRecommend(uid, page)
		if err != nil {
			return nil, errors.New("获取推荐失败Orz")
		}
		posters, err := s.getPosters(ids)
		if err != nil {
			return nil, errors.New("获取帖子失败Orz")
		}
		if len(posters) < int(config.Get().RecommendPageSize) {
			var posters_ []database.Poster
			if err := database.DB.Order("RANDOM()").Where("public = true").Limit(int(config.Get().RecommendPageSize) - len(posters)).Find(&posters_).Error; err != nil {
				return nil, errors.New("获取帖子失败Orz")
			}
			posters = append(posters, posters_...)
		}
		return s.posters2PosterAPIs(posters), nil
	} else if mode == "hot" {
		popular, err := gorse.GetPopular(page)
		if err != nil {
			return nil, errors.New("获取热榜失败Orz")
		}
		posters, err := s.getPosters(popular)
		if err != nil {
			return nil, errors.New("获取帖子失败Orz")
		}
		return s.posters2PosterAPIs(posters), nil
	}
	// follow/search模式
	var orders []string
	switch order {
	case "like":
		orders = append(orders, "like_num:desc")
	default: //默认new
		orders = append(orders, "create_time:desc")
	}
	var filters []string
	if mode == "follow" {
		keyword = ""
		var follows []database.Follow
		database.DB.Where("uid = ?", uid).Find(&follows)
		uidFilter := "uid IN [ "
		for _, follow := range follows {
			uidFilter += fmt.Sprintf("%v,", follow.FollowUid)
		}
		uidFilter = uidFilter[:len(uidFilter)-1] + "]"
		filters = append(filters, uidFilter)
		filters = append(filters, "anonymous = false")
		filters = append(filters, "public = true")
	} else { //搜索模式
		if uid != 0 {
			filters = append(filters, "uid ="+fmt.Sprintf("%v", uid))
		}
		if noAnonymous {
			filters = append(filters, "anonymous = false")
		}
		if onlyPublic {
			filters = append(filters, "public = true")
		}
	}
	var posters []database.Poster
	err := search.Search(&posters, s.GetObjType(), keyword, page, config.Get().PostPageSize, orders, filters)
	if err != nil {
		return nil, errors.New("搜索失败Orz")
	}
	return s.posters2PosterAPIs(posters), nil
}

// Delete 删除帖子
func (s *PosterService) Delete(id, uid uint, admin bool) error {
	poster, err := s.getPoster(id)
	if err != nil {
		return err
	}
	if poster.Uid != uid && !admin {
		return errors.New("没有删除权限Orz")
	}
	if err := database.DB.Delete(&poster).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	go func() {
		search.Delete(s.GetObjType(), []string{fmt.Sprintf("%v", poster.ID)})
		gorse.DeletePoster(id)
	}()
	return nil
}

// GetClaims 获取声明列表
func (s *PosterService) GetClaims() []database.Claim {
	var claims = make([]database.Claim, 0)
	for _, claim := range database.ClaimMap {
		claims = append(claims, claim)
	}
	return claims
}
