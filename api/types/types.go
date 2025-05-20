package types

import (
	"BIT101-GO/database"
	"time"
)

// Image 图片模块

// ImageAPI 用于API的图片结构 区别于数据库
type ImageAPI struct {
	Mid    string `json:"mid"`
	Url    string `json:"url"`
	LowURL string `json:"low_url"`
}

// User 用户模块

// 枚举用户类型(需要与数据库中定义一致)
type IdentityType uint

// 用户身份类型常量
const (
	IdentityNormal IdentityType = iota
	IdentityAdmin
	IdentitySuper
	IdentityOrganization
	IdentityClub
	IdentityNGO
	IdentityRobot
	IdentityBig
)

// UserAPI 用户信息
type UserAPI struct {
	ID         int               `json:"id"`
	CreateTime time.Time         `json:"create_time"`
	Nickname   string            `json:"nickname"`
	Avatar     ImageAPI          `json:"avatar"`
	Motto      string            `json:"motto"`
	Identity   database.Identity `json:"identity"`
}

// FollowAPI 关注信息
type FollowAPI struct {
	FollowingNum int64 `json:"following_num"` // 关注数
	FollowerNum  int64 `json:"follower_num"`  // 粉丝数
	Following    bool  `json:"following"`     // 是否被我关注
	Follower     bool  `json:"follower"`      // 是否关注我
}

// UserInfo 用户详细信息
type UserInfo struct {
	UserAPI UserAPI `json:"user"`
	FollowAPI
	Own bool `json:"own"` // 是否是自己
}

// Message 消息模块

// MessageType 消息类型
type MessageType string

const (
	MessageTypeFollow       MessageType = "follow"
	MessageTypeComment      MessageType = "comment"
	MessageTypeLike         MessageType = "like"
	MessageTypeSystem       MessageType = "system"
	MessageTypeSubscription MessageType = "subscription"
)

var MessageTypeMap = map[MessageType]*struct{}{
	MessageTypeFollow:  nil,
	MessageTypeComment: nil,
	MessageTypeLike:    nil,
	MessageTypeSystem:  nil,
}

// MessageAPI 消息列表的元素
type MessageAPI struct {
	ID         uint      `json:"id"`        // 消息ID
	FromUser   UserAPI   `json:"from_user"` // 发送者
	LinkObj    string    `json:"link_obj"`  // 链接对象
	Obj        string    `json:"obj"`       // 关联对象
	Text       string    `json:"text"`      // 内容
	UpdateTime time.Time `json:"update_time"`
}

// Webpush 消息推送
type WebpushMessage struct {
	Data      []byte `json:"data"`
	Badge     string `json:"badge"`
	Icon      string `json:"icon"`
	Timestamp int64  `json:"timestamp"`
}

// Manage 管理模块

// ReportAPI 举报信息
type ReportAPI struct {
	CreatedTime string              `json:"created_time"`
	ID          int                 `json:"id"`
	Obj         string              `json:"obj"`
	ReportType  database.ReportType `json:"report_type"`
	Status      int                 `json:"status"` // 0为未处理 1为举报成功 2为举报失败
	Text        string              `json:"text"`
	User        UserAPI             `json:"user"`
}

// BanAPI 封禁信息
type BanAPI struct {
	ID   uint      `json:"id"`
	Time time.Time `json:"time"`
	User UserAPI   `json:"user"`
}

// Reaction 交互模块

// LikeAPI 点赞信息
type LikeAPI struct {
	Like    bool `json:"like"`     // 是否点赞
	LikeNum uint `json:"like_num"` // 点赞数
}

// CommentAPI 评论信息
type CommentAPI struct {
	database.Comment
	Like      bool         `json:"like"`       // 是否点赞
	Own       bool         `json:"own"`        // 是否是自己的评论
	ReplyUser UserAPI      `json:"reply_user"` // 回复的用户
	Sub       []CommentAPI `json:"sub"`        // 子评论
	User      UserAPI      `json:"user"`       // 评论用户
	Images    []ImageAPI   `json:"images"`     // 图片
}

// Course 课程模块

// CourseAPI 课程信息
type CourseAPI struct {
	database.Course
	Like         bool                       `json:"like"`         // 是否点赞
	Subscription database.SubscriptionLevel `json:"subscription"` // 订阅级别
}

// CourseHistoryAPI 课程历史信息
type CourseHistoryAPI struct {
	Term       string  `json:"term"`        //学期
	AvgScore   float64 `json:"avg_score"`   //均分
	MaxScore   float64 `json:"max_score"`   //最高分
	StudentNum uint    `json:"student_num"` //学习人数
}

// Paper 文章模块

// PpaerAPI 文章信息
type PaperAPI struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Intro      string    `json:"intro"`
	LikeNum    uint      `json:"like_num"`
	CommentNum uint      `json:"comment_num"`
	UpdateTime time.Time `json:"update_time"`
}

// PaperInfo 文章详细信息
type PaperInfo struct {
	database.Paper
	UpdateUser   UserAPI                    `json:"update_user"`
	Like         bool                       `json:"like"`
	Subscription database.SubscriptionLevel `json:"subscription"`
	Own          bool                       `json:"own"`
}

// Poster 帖子模块

// PosterAPI 帖子信息
type PosterAPI struct {
	database.Poster
	User   UserAPI        `json:"user"`
	Images []ImageAPI     `json:"images"`
	Tags   []string       `json:"tags"`
	Claim  database.Claim `json:"claim"`
}

// PosterInfo 帖子详细信息
type PosterInfo struct {
	database.Poster
	User         UserAPI                    `json:"user"`
	Images       []ImageAPI                 `json:"images"`
	Tags         []string                   `json:"tags"`
	Claim        database.Claim             `json:"claim"`
	Like         bool                       `json:"like"`
	Own          bool                       `json:"own"`
	Subscription database.SubscriptionLevel `json:"subscription"`
}

// Gorse 推荐模块

type FeedbackType string

const (
	FeedbackTypeRead    FeedbackType = "read"
	FeedbackTypeLike    FeedbackType = "like"
	FeedbackTypeComment FeedbackType = "comment"
	FeedbackTypeStay    FeedbackType = "stay"
)

// Subscription 订阅模块

// SubscriptionAPI 订阅API
type SubscriptionAPI struct {
	ID               uint      `json:"id"`                // 订阅编号
	SubscriptionTime time.Time `json:"subscription_time"` // 订阅时间
	Obj              string    `json:"obj"`               // 订阅对象
	Level            int       `json:"level"`             // 订阅级别 0:silent | 1:update | 2:comment
	Text             string    `json:"text"`              // 简介
}
