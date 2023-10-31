/**
* @author:YHCnb
* Package:
* @date:2023/10/18 21:17
* Description:
 */
package gorse

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"context"
	"github.com/zhenghaoz/gorse/client"
	"strconv"
	"strings"
	"time"
)

var gorse *client.GorseClient

// InitUserAndItem 初始化用户和item(仅使用一次)
func InitUserAndItem() {
	// 从db数据库取出所有user
	var users []database.User
	database.DB.Find(&users)
	err := InsertUsers(users)
	if err != nil {
		println("初始化users失败:", err.Error())
		return
	}
	// 从db数据库取出所有poster
	var posters []database.Poster
	database.DB.Find(&posters)
	err = InsertPosters(posters)
	if err != nil {
		println("初始化items失败:", err.Error())
		return
	}
}

// Sync 同步
func Sync(time_after time.Time) {
	users := []database.User{}
	users_update := []database.User{}
	q := database.DB.Model(database.User{}).Unscoped().Where("updated_at > ?", time_after).Or("deleted_at > ?", time_after)
	if err := q.Find(&users).Error; err != nil {
		println("同步users失败:", err.Error())
	}
	for _, v := range users {
		if v.DeletedAt.Valid {
			DeleteUser(strconv.Itoa(int(v.ID)))
		} else {
			users_update = append(users_update, v)
		}
	}
	go func() {
		if err := InsertUsers(users_update); err != nil {
			println("同步users失败:", err.Error())
		}
	}()

	posters := []database.Poster{}
	posters_update := []database.Poster{}
	q = database.DB.Model(database.Poster{}).Unscoped().Where("updated_at > ?", time_after).Or("deleted_at > ?", time_after)
	if err := q.Find(&posters).Error; err != nil {
		println("同步items失败:", err.Error())
	}
	for _, v := range posters {
		if v.DeletedAt.Valid {
			DeletePoster(strconv.Itoa(int(v.ID)))
		} else {
			posters_update = append(posters_update, v)
		}
	}
	go func() {
		if err := InsertPosters(posters_update); err != nil {
			println("同步items失败:", err.Error())
		}
	}()
}

// User2GorseUser user转换为gorseUser
func User2GorseUser(user database.User) client.User {
	return client.User{
		UserId:    strconv.Itoa(int(user.ID)),
		Labels:    []string{},
		Subscribe: []string{},
		Comment:   user.Nickname,
	}
}

// InsertUser 插入用户
func InsertUser(user database.User) error {
	return InsertUsers([]database.User{user})
}

// InsertUsers 插入多个用户
func InsertUsers(users []database.User) error {
	var gorseUsers []client.User
	for _, user := range users {
		gorseUsers = append(gorseUsers, User2GorseUser(user))
	}
	_, err := gorse.InsertUsers(context.Background(), gorseUsers)
	return err
}

// UpdateUser 更新用户
func UpdateUser(user database.User) error {
	gorseUser := User2GorseUser(user)
	_, err := gorse.UpdateUser(context.Background(), gorseUser.UserId, client.UserPatch{
		Labels:    gorseUser.Labels,
		Subscribe: gorseUser.Subscribe,
		Comment:   &gorseUser.Comment,
	})
	return err
}

// DeleteUser 删除用户
func DeleteUser(id string) error {
	_, err := gorse.DeleteUser(context.Background(), id)
	return err
}

// Poster2GorseItem post转换为gorseItem
func Poster2GorseItem(post database.Poster) client.Item {
	return client.Item{
		ItemId:     strconv.Itoa(int(post.ID)),
		IsHidden:   !post.Public,
		Labels:     strings.Split(post.Tags, " "),
		Categories: []string{},
		Timestamp:  post.EditAt.String(),
		Comment:    post.Title,
	}
}

// InsertPoster 插入poster
func InsertPoster(post database.Poster) error {
	return InsertPosters([]database.Poster{post})
}

// InsertPosters 插入多个poster
func InsertPosters(posts []database.Poster) error {
	var gorseItems []client.Item
	for _, post := range posts {
		gorseItems = append(gorseItems, Poster2GorseItem(post))
	}
	_, err := gorse.InsertItems(context.Background(), gorseItems)
	return err
}

// UpdatePoster 更新poster
func UpdatePoster(post database.Poster) error {
	gorseItem := Poster2GorseItem(post)
	_, err := gorse.UpdateItem(context.Background(), gorseItem.ItemId, client.ItemPatch{
		IsHidden:   &gorseItem.IsHidden,
		Labels:     gorseItem.Labels,
		Categories: gorseItem.Categories,
		Timestamp:  &post.EditAt,
		Comment:    &gorseItem.Comment,
	})
	return err
}

// DeletePoster 删除poster
func DeletePoster(id string) error {
	_, err := gorse.DeleteItem(context.Background(), id)
	return err
}

// InsertFeedback 插入反馈
func InsertFeedback(feedback client.Feedback) error {
	return InsertFeedbacks([]client.Feedback{feedback})
}

// InsertFeedbacks 插入多条反馈
func InsertFeedbacks(feedbacks []client.Feedback) error {
	_, err := gorse.InsertFeedback(context.Background(), feedbacks)
	return err
}

// DeleteFeedback 删除反馈
func DeleteFeedback(feedbackType, userId, itemId string) error {
	_, err := gorse.DelFeedback(context.Background(), feedbackType, userId, itemId)
	return err
}

// GetPopular 获取popular
func GetPopular(page uint) ([]string, error) {
	limit := int(config.Config.PostPageSize)
	score, err := gorse.GetItemPopular(context.Background(), "", limit, int(page)*limit)
	if err != nil {
		return nil, err
	}
	var popular []string
	for _, item := range score {
		popular = append(popular, item.Id)
	}
	return popular, nil
}

// GetRecommend 获取recommend
func GetRecommend(userid string, page uint) ([]string, error) {
	limit := int(config.Config.RecommendPageSize)
	recommend, err := gorse.GetItemRecommend(context.Background(), userid, []string{}, "read", "", limit, int(page)*limit)
	if err != nil {
		return nil, err
	}
	return recommend, nil
}

// Init 初始化
func Init() {
	gorse = client.NewGorseClient("http://127.0.0.1:8088", "BIT101")
	InitUserAndItem()
}
