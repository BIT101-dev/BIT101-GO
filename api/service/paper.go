/*
 * @Author: flwfdd
 * @Date: 2025-03-11 11:12:41
 * @LastEditTime: 2025-03-11 12:04:19
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/search"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ types.PaperService = (*PaperService)(nil)

type PaperService struct {
	userSvc     types.UserService
	reactionSvc types.ReactionService
}

func NewPaperService(userSvc types.UserService, reactionSvc types.ReactionService) *PaperService {
	s := PaperService{userSvc: userSvc, reactionSvc: reactionSvc}
	types.RegisterObjHandler(&s)
	return &s
}

/* ObjHandler */

// IsExist 判断文章是否存在
func (s *PaperService) IsExist(id uint) bool {
	_, err := s.getPaper(id)
	return err == nil
}

// GetObjType 获取文章类型
func (s *PaperService) GetObjType() string {
	return "paper"
}

// GetUid 获取文章作者
func (s *PaperService) GetUid(id uint) (uint, error) {
	return 0, errors.ErrUnsupported
}

// LikeHandler 点赞文章
func (s *PaperService) LikeHandler(tx *gorm.DB, id uint, delta int, uid uint) (likeNum uint, err error) {
	var paper database.Paper
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&paper, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if paper.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	paper.LikeNum = uint(int(paper.LikeNum) + delta)
	if err := tx.Save(&paper).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		search.Update(s.GetObjType(), paper)
	}()
	return paper.LikeNum, nil
}

// CommentHandler 评论文章
func (s *PaperService) CommentHandler(tx *gorm.DB, id uint, comment database.Comment, delta int, uid uint) (commentNum uint, err error) {
	var paper database.Paper
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&paper, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if paper.ID == 0 {
		return 0, errors.New("文章不存在Orz")
	}
	paper.CommentNum = uint(int(paper.CommentNum) + delta)
	if err := tx.Save(&paper).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		search.Update(s.GetObjType(), paper)
	}()
	return paper.CommentNum, nil
}

// getPaper 获取文章
func (s *PaperService) getPaper(id uint) (paper database.Paper, err error) {
	if err := database.DB.Limit(1).Find(&paper, "id = ?", id).Error; err != nil {
		return database.Paper{}, errors.New("数据库错误Orz")
	}
	if paper.ID == 0 {
		return database.Paper{}, errors.New("文章不存在Orz")
	}
	return paper, nil
}

// Get 获取文章
func (s *PaperService) Get(id, uid uint, admin bool) (types.PaperInfo, error) {
	paper, err := s.getPaper(id)
	if err != nil {
		return types.PaperInfo{}, err
	}

	var updateUid int
	if paper.Anonymous {
		updateUid = -1
	} else {
		updateUid = int(paper.UpdateUid)
	}
	updateUser, err := s.userSvc.GetUserAPI(updateUid)
	if err != nil {
		return types.PaperInfo{}, errors.New("获取编辑用户信息失败Orz")
	}
	paper.CreateUid = 0
	paper.UpdateUid = 0

	like := s.reactionSvc.CheckLike(fmt.Sprintf("%s%v", s.GetObjType(), paper.ID), uid)

	own := paper.CreateUid == uid || admin
	paper.UpdatedAt = paper.EditAt
	return types.PaperInfo{
		Paper:      paper,
		UpdateUser: updateUser,
		Like:       like,
		Own:        own,
	}, nil
}

// 添加到编辑记录
func pushHistory(tx *gorm.DB, paper database.Paper) error {
	history := database.PaperHistory{
		Pid:       paper.ID,
		Title:     paper.Title,
		Intro:     paper.Intro,
		Content:   paper.Content,
		Uid:       paper.UpdateUid,
		Anonymous: paper.Anonymous,
	}
	if err := database.DB.Create(&history).Error; err != nil {
		return err
	}
	return nil
}

// Create 创建文章
func (s *PaperService) Create(title, intro, content string, anonymous, publicEdit bool, uid uint) (uint, error) {
	var paperID uint
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var paper = database.Paper{
			Title:      title,
			Intro:      intro,
			Content:    content,
			CreateUid:  uid,
			UpdateUid:  uid,
			Anonymous:  anonymous,
			PublicEdit: publicEdit,
		}
		if err := tx.Create(&paper).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if err := pushHistory(tx, paper); err != nil {
			return errors.New("编辑日志错误Orz")
		}
		go func() {
			search.Update(s.GetObjType(), paper)
		}()
		paperID = paper.ID
		return nil
	}); err != nil {
		return 0, err
	}
	return paperID, nil
}

// Edit 编辑文章
func (s *PaperService) Edit(id uint, title, intro, content string, anonymous, publicEdit bool, lastTime float64, uid uint, admin bool) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		paper, err := s.getPaper(id)
		if err != nil {
			return err
		}
		if paper.CreateUid != uid && !paper.PublicEdit && !admin {
			return errors.New("没有编辑权限Orz")
		}
		if paper.EditAt.Unix() > int64(lastTime) {
			return errors.New("请基于最新版本编辑Orz")
		}

		paper.Title = title
		paper.Intro = intro
		paper.Content = content
		paper.Anonymous = anonymous
		paper.UpdateUid = uid
		paper.EditAt = time.Now()
		if paper.CreateUid == uid || admin {
			paper.PublicEdit = publicEdit
		}
		if err := tx.Save(&paper).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		pushHistory(tx, paper)
		go func() {
			search.Update(s.GetObjType(), paper)
		}()
		return nil
	})
}

// GetList 获取文章列表
func (s *PaperService) GetList(keyword string, order string, page uint) ([]types.PaperAPI, error) {
	// 排序
	var orders []string
	if order != "rand" {
		if order == "like" {
			orders = append(orders, "like_num:desc")
		} else { //默认new
			orders = append(orders, "update_time:desc")
		}
	}

	var papers []database.Paper
	err := search.Search(&papers, "paper", keyword, page, config.Get().PaperPageSize, orders, nil)
	if err != nil {
		return nil, errors.New("搜索失败Orz")
	}
	paperList := make([]types.PaperAPI, 0)
	for _, paper := range papers {
		paperList = append(paperList, types.PaperAPI{
			ID:         paper.ID,
			Title:      paper.Title,
			Intro:      paper.Intro,
			LikeNum:    paper.LikeNum,
			CommentNum: paper.CommentNum,
			UpdateTime: paper.EditAt,
		})
	}
	return paperList, nil
}

// Delete 删除文章
func (s *PaperService) Delete(id, uid uint, admin bool) error {
	paper, err := s.getPaper(id)
	if err != nil {
		return err
	}
	if paper.CreateUid != uid && !admin {
		return errors.New("没有删除权限Orz")
	}
	if err := database.DB.Delete(&paper).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	go func() {
		search.Delete(s.GetObjType(), []string{strconv.Itoa(int(paper.ID))})
	}()
	return nil
}
