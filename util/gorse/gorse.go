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
	// 从db数据库取出所有post
	var posts []database.Post
	database.DB.Find(&posts)
	err = InsertPosts(posts)
	if err != nil {
		println("初始化items失败:", err.Error())
		return
	}
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
	if _, err := gorse.InsertUsers(context.Background(), gorseUsers); err != nil {
		return err
	}
	return nil
}

// UpdateUser 更新用户
func UpdateUser(user database.User) error {
	gorseUser := User2GorseUser(user)
	_, err := gorse.UpdateUser(context.Background(), gorseUser.UserId, client.UserPatch{
		Labels:    gorseUser.Labels,
		Subscribe: gorseUser.Subscribe,
		Comment:   &gorseUser.Comment,
	})
	if err != nil {
		return err
	}
	return nil
}

// Post2GorseItem post转换为gorseItem
func Post2GorseItem(post database.Post) client.Item {
	return client.Item{
		ItemId:     strconv.Itoa(int(post.ID)),
		IsHidden:   !post.Public,
		Labels:     strings.Split(post.Tags, " "),
		Categories: []string{},
		Timestamp:  post.EditAt.String(),
		Comment:    post.Title,
	}
}

// InsertPost 插入post
func InsertPost(post database.Post) error {
	return InsertPosts([]database.Post{post})
}

// InsertPosts 插入多个post
func InsertPosts(posts []database.Post) error {
	var gorseItems []client.Item
	for _, post := range posts {
		gorseItems = append(gorseItems, Post2GorseItem(post))
	}
	if _, err := gorse.InsertItems(context.Background(), gorseItems); err != nil {
		return err
	}
	return nil
}

// UpdatePost 更新post
func UpdatePost(post database.Post) error {
	gorseItem := Post2GorseItem(post)
	_, err := gorse.UpdateItem(context.Background(), gorseItem.ItemId, client.ItemPatch{
		IsHidden:   &gorseItem.IsHidden,
		Labels:     gorseItem.Labels,
		Categories: gorseItem.Categories,
		Timestamp:  &post.EditAt,
		Comment:    &gorseItem.Comment,
	})
	if err != nil {
		return err
	}
	return nil
}

// InsertFeedback 插入反馈
func InsertFeedback(feedback client.Feedback) error {
	return InsertFeedbacks([]client.Feedback{feedback})
}

// InsertFeedbacks 插入多条反馈
func InsertFeedbacks(feedbacks []client.Feedback) error {
	if _, err := gorse.InsertFeedback(context.Background(), feedbacks); err != nil {
		return err
	}
	return nil
}

// GetPopular 获取popular
func GetPopular(userid string, page uint) ([]string, error) {
	limit := int(config.Config.PostPageSize)
	score, err := gorse.GetItemPopular(context.Background(), userid, limit, int(page)*limit)
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
	limit := int(config.Config.PostPageSize)
	recommend, err := gorse.GetItemRecommend(context.Background(), userid, []string{}, "", "", limit, int(page)*limit)
	if err != nil {
		return nil, err
	}
	return recommend, nil
}

// Init 初始化
func Init() {
	gorse = client.NewGorseClient("http://127.0.0.1:8088", "BIT101")

	//InitUserAndItem()
	//
	//gorse.InsertFeedback(context.Background(), []client.Feedback{
	//	{FeedbackType: "like", UserId: "bob", ItemId: "vuejs:vue", Timestamp: "2022-02-24"},
	//	{FeedbackType: "like", UserId: "bob", ItemId: "d3:d3", Timestamp: "2022-02-25"},
	//	{FeedbackType: "like", UserId: "bob", ItemId: "dogfalo:materialize", Timestamp: "2022-02-26"},
	//	{FeedbackType: "like", UserId: "bob", ItemId: "mozilla:pdf.js", Timestamp: "2022-02-27"},
	//	{FeedbackType: "like", UserId: "bob", ItemId: "moment:moment", Timestamp: "2022-02-28"},
	//})

	// Get recommendation.
	//recommend, err := gorse.GetItemRecommend(context.Background(), "bob", []string{}, "", "", 10, 0)
	//if err != nil {
	//	println("err.............")
	//	return
	//}
	//for _, item := range recommend {
	//	println(item)
	//}
}
