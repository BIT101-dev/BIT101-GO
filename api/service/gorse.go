/**
* @author:YHCnb
* Package:
* @date:2023/10/18 21:17
* Description:
 */
package service

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/zhenghaoz/gorse/client"
)

type GorseService struct {
	client        *client.GorseClient
	syncInterval  time.Duration
	syncBatchSize int
}

func NewGorseService() *GorseService {
	s := GorseService{}
	s.init()
	return &s
}

// SyncUsers 同步用户
func (s *GorseService) syncUsers(users []database.User) {
	userMap := make(map[string]database.User)
	for _, user := range users {
		userMap[strconv.Itoa(int(user.ID))] = user
	}
	cursor := ""
	for {
		gorseUsers, err := s.client.GetUsers(context.Background(), cursor, s.syncBatchSize)
		if err != nil {
			slog.Error("SyncUsers", "opt", "GetUsers", "err", err.Error())
			return
		}
		for _, gorseUser := range gorseUsers.Users {
			if _, ok := userMap[gorseUser.UserId]; !ok {
				s.deleteUser(gorseUser.UserId)
			}
		}
		if gorseUsers.Cursor == "" {
			break
		}
		cursor = gorseUsers.Cursor
	}
	err := s.InsertUsers(users)
	if err != nil {
		slog.Error("SyncUsers", "opt", "InsertUsers", "err", err.Error())
		return
	}
}

// SyncPosters 同步posters
func (s *GorseService) syncPosters(posters []database.Poster) {
	posterMap := make(map[string]database.Poster)
	for _, poster := range posters {
		posterMap[strconv.Itoa(int(poster.ID))] = poster
	}
	cursor := ""
	for {
		gorseItems, err := s.client.GetItems(context.Background(), cursor, s.syncBatchSize)
		if err != nil {
			slog.Error("SyncPosters", "opt", "GetItems", "err", err.Error())
			return
		}
		for _, gorseItem := range gorseItems.Items {
			if _, ok := posterMap[gorseItem.ItemId]; !ok {
				s.deletePoster(gorseItem.ItemId)
			}
		}
		if gorseItems.Cursor == "" {
			break
		}
		cursor = gorseItems.Cursor
	}
	err := s.InsertPosters(posters)
	if err != nil {
		slog.Error("SyncPosters", "opt", "InsertPosters", "err", err.Error())
		return
	}
}

// sync 同步
func (s *GorseService) sync() {
	defer (func() {
		if err := recover(); err != nil {
			slog.Error("gorse: sync failed", "err", err)
		} else {
			slog.Info("gorse: sync success")
		}
	})()

	// 从数据库取出所有user
	var users []database.User
	database.DB.Find(&users)
	s.syncUsers(users)

	// 从数据库取出所有poster
	var posters []database.Poster
	database.DB.Find(&posters)
	s.syncPosters(posters)
}

// user2Gorse user转换为gorseUser
func (s *GorseService) user2Gorse(user database.User) client.User {
	return client.User{
		UserId:    strconv.Itoa(int(user.ID)),
		Labels:    []string{},
		Subscribe: []string{},
		Comment:   user.Nickname,
	}
}

// InsertUser 插入用户
func (s *GorseService) InsertUser(user database.User) error {
	return s.InsertUsers([]database.User{user})
}

// InsertUsers 插入多个用户
func (s *GorseService) InsertUsers(users []database.User) error {
	var gorseUsers []client.User
	for _, user := range users {
		gorseUsers = append(gorseUsers, s.user2Gorse(user))
	}
	_, err := s.client.InsertUsers(context.Background(), gorseUsers)
	return err
}

// UpdateUser 更新用户
func (s *GorseService) UpdateUser(user database.User) error {
	gorseUser := s.user2Gorse(user)
	_, err := s.client.UpdateUser(context.Background(), gorseUser.UserId, client.UserPatch{
		Labels:    gorseUser.Labels,
		Subscribe: gorseUser.Subscribe,
		Comment:   &gorseUser.Comment,
	})
	return err
}

// DeleteUser 删除用户
func (s *GorseService) deleteUser(id string) error {
	_, err := s.client.DeleteUser(context.Background(), id)
	return err
}

// Poster2GorseItem poster转换为gorseItem
func (s *GorseService) poster2Gorse(poster database.Poster) client.Item {
	return client.Item{
		ItemId:     strconv.Itoa(int(poster.ID)),
		IsHidden:   !poster.Public,
		Labels:     common.Spilt(poster.Tags),
		Categories: []string{},
		Timestamp:  poster.CreatedAt.String(),
		Comment:    poster.Title,
	}
}

// InsertPoster 插入poster
func (s *GorseService) InsertPoster(post database.Poster) error {
	return s.InsertPosters([]database.Poster{post})
}

// InsertPosters 插入多个poster
func (s *GorseService) InsertPosters(posts []database.Poster) error {
	var gorseItems []client.Item
	for _, post := range posts {
		gorseItems = append(gorseItems, s.poster2Gorse(post))
	}
	_, err := s.client.InsertItems(context.Background(), gorseItems)
	return err
}

// UpdatePoster 更新poster
func (s *GorseService) UpdatePoster(post database.Poster) error {
	gorseItem := s.poster2Gorse(post)
	// Convert string timestamp to time.Time object
	t, err := time.Parse(time.RFC3339, gorseItem.Timestamp)
	if err != nil {
		return err
	}
	_, err = s.client.UpdateItem(context.Background(), gorseItem.ItemId, client.ItemPatch{
		IsHidden:   &gorseItem.IsHidden,
		Labels:     gorseItem.Labels,
		Categories: gorseItem.Categories,
		Timestamp:  &t,
		Comment:    &gorseItem.Comment,
	})
	return err
}

// DeletePoster 删除poster
func (s *GorseService) deletePoster(id string) error {
	_, err := s.client.DeleteItem(context.Background(), id)
	return err
}

// DeletePoster 删除poster
func (s *GorseService) DeletePoster(id uint) error {
	return s.deletePoster(strconv.Itoa(int(id)))
}

// InsertFeedback 插入反馈
func (s *GorseService) InsertFeedback(typ types.FeedbackType, userId, itemId string) error {
	_, err := s.client.InsertFeedback(context.Background(), []client.Feedback{
		{
			FeedbackType: string(typ),
			UserId:       userId,
			ItemId:       itemId,
			Timestamp:    time.Now().String(),
		},
	})
	return err
}

// DeleteFeedback 删除反馈
func (s *GorseService) DeleteFeedback(typ types.FeedbackType, userId, itemId string) error {
	_, err := s.client.DelFeedback(context.Background(), string(typ), userId, itemId)
	return err
}

// GetPopular 获取popular
func (s *GorseService) GetPopular(page uint) ([]uint, error) {
	limit := int(config.Get().PostPageSize)
	posters, err := s.client.GetItemPopular(context.Background(), "", limit, int(page)*limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0)
	for _, item := range posters {
		id, err := strconv.Atoi(item.Id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uint(id))
	}
	return ids, nil
}

// GetRecommend 获取recommend
func (s *GorseService) GetRecommend(uid uint, page uint) ([]uint, error) {
	limit := int(config.Get().RecommendPageSize)
	posters, err := s.client.GetItemRecommend(context.Background(), strconv.Itoa(int(uid)), []string{}, string(types.FeedbackTypeRead), "", limit, int(page)*limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0)
	for _, item := range posters {
		id, err := strconv.Atoi(item)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uint(id))
	}
	return ids, nil
}

// init 初始化客户端并同步数据
func (s *GorseService) init() {
	// 初始化client
	s.client = client.NewGorseClient("http://127.0.0.1:8088", "BIT101")
	s.syncInterval = time.Duration(config.Get().SyncInterval) * time.Second
	s.syncBatchSize = 1000

	// 定时同步
	go func() {
		for {
			go s.sync()
			time.Sleep(s.syncInterval)
		}
	}()
}
