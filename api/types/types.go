package types

import (
	"BIT101-GO/database"
	"BIT101-GO/pkg/webvpn"
	"time"

	"github.com/SherClockHolmes/webpush-go"
)

// Image 图片模块

// ImageAPI 用于API的图片结构 区别于数据库
type ImageAPI struct {
	Mid    string `json:"mid"`
	Url    string `json:"url"`
	LowURL string `json:"low_url"`
}

// ImageService 图片模块服务接口
type ImageService interface {
	GetImageAPI(mid string) ImageAPI              // 通过mid生成ImageAPI
	GetImageAPIList(mids []string) []ImageAPI     // 通过mids生成ImageAPI数组
	Mid2Url(mid string) string                    // 将图片mid转换为url
	CheckMid(mid string) bool                     // 检查图片mid是否存在
	CheckMids(mids []string) bool                 // 批量检查图片mid是否存在
	Save(uid uint, data []byte) (ImageAPI, error) // 保存图片
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

// UserService 用户模块服务接口
type UserService interface {
	// 用户信息相关
	GetAnonymousUserAPI() UserAPI                                // 获取匿名用户信息
	GetAnonymousName(uid uint, obj string) string                // 获取匿名用户名
	GetUserAPI(uid int) (UserAPI, error)                         // 获取用户信息
	GetUserAPIList(uid_list []int) ([]UserAPI, error)            // 批量获取用户信息
	GetUserAPIMap(uid_map map[int]bool) (map[int]UserAPI, error) // 批量获取用户信息
	GetInfo(uid int, myUid uint) (UserInfo, error)               // 获取用户详细信息
	OldGetInfo(uid int) (UserAPI, error)                         // 获取用户信息(旧)
	SetInfo(uid int, nickname, avatarMid, motto string) error    // 修改用户信息

	// 认证相关
	Login(sid string, pwd string) (fakeCookie string, err error)                                                                 // 登录
	WebvpnVerifyInit(sid string) (data webvpn.InitLoginReturn, captcha string, err error)                                        // Webvpn登录初始化
	WebvpnVerify(sid string, pwd string, execution string, cookie string, captcha string) (token string, code string, err error) // Webvpn登录验证
	MailVerify(sid string) (token string, err error)                                                                             // 邮箱验证
	Register(token string, code string, pwd string, loginMode bool) (fakeCookie string, err error)                               // 注册

	// 关注相关
	GetFollowAPI(followUid, myUid uint) (FollowAPI, error) // 获取关注信息
	Follow(followUid int, myUid uint) error                // 关注
	GetFollowList(uid uint, page uint) ([]UserAPI, error)  // 获取关注列表
	GetFansList(uid uint, page uint) ([]UserAPI, error)    // 获取粉丝列表
}

// Message 消息模块

// MessageType 消息类型
type MessageType string

const (
	MessageTypeFollow  MessageType = "follow"
	MessageTypeComment MessageType = "comment"
	MessageTypeLike    MessageType = "like"
	MessageTypeSystem  MessageType = "system"
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

// MessageService 消息模块服务接口
type MessageService interface {
	// 内部消息
	Send(fromUid int, toUid uint, objID string, typ MessageType, linkObjID string, text string) error // 发送消息
	GetList(uid uint, typ MessageType, lastID uint) ([]MessageAPI, error)                             // 获取消息列表
	GetUnreadNum(uid uint) (uint, error)                                                              // 获取总未读消息数
	GetUnreadNums(uid uint) (map[string]uint, error)                                                  // 获取各类消息未读数

	// Webpush
	WebpushSubscribe(sub webpush.Subscription, uid uint) error // 订阅Webpush
	WebpushSend(uid uint, content []byte) error                // 推送Webpush消息
	WebpushGetRequestPubkey() string                           // 获取Webpush请求公钥
	WebpushSendToSubscription(sub database.WebPushSubscription, message []byte) error
	WebpushCancelRegister(sub webpush.Subscription, uid uint) error
}

// Webpush 消息推送
type WebpushMessage struct {
	Data      []byte `json:"data"`
	Badge     string `json:"badge"`
	Icon      string `json:"icon"`
	Timestamp int64  `json:"timestamp"`
}

// Variable 变量模块
type VariableService interface {
	Get(objID string) (string, error) // 获取变量
	Set(objID, data string) error     // 设置变量
}

// Manage 管理模块

// ManageService 管理模块服务接口
type ManageService interface {
	GetReportTypes() []database.ReportType                                        // 获取举报类型列表
	Report(reporter uint, objID string, typeId uint, text string) error           // 举报
	UpdateReportStatus(id uint, status int) error                                 // 更新举报状态
	GetReports(uid int, objID string, status int, page uint) ([]ReportAPI, error) // 获取举报列表
	Ban(uid uint, time time.Time) error                                           // 封禁
	GetBans(uid int, page uint) ([]BanAPI, error)                                 // 获取封禁列表
}

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

// ReactionService 交互模块服务接口
type ReactionService interface {
	Like(obj Obj, uid uint) (LikeAPI, error)                                                                                                      // 点赞
	CheckLike(objID string, uid uint) bool                                                                                                        // 检查是否点赞
	CheckLikeMap(objIDMap map[string]bool, uid uint) (map[string]bool, error)                                                                     // 批量检查是否点赞
	Comment(obj Obj, text string, anonymous bool, replyUid int, replyObj Obj, rate uint, mids []string, uid uint, admin bool) (CommentAPI, error) // 评论
	GetComments(obj Obj, order string, page uint, uid uint, admin bool) ([]CommentAPI, error)                                                     // 获取评论列表
	DeleteComment(id uint, uid uint, admin bool) error                                                                                            // 删除评论
	Stay(obj Obj, uid uint) error                                                                                                                 // 停留上报
}

// Course 课程模块

// CourseAPI 课程信息
type CourseAPI struct {
	database.Course
	Like bool `json:"like"` // 是否点赞
}

// CourseHistoryAPI 课程历史信息
type CourseHistoryAPI struct {
	Term       string  `json:"term"`        //学期
	AvgScore   float64 `json:"avg_score"`   //均分
	MaxScore   float64 `json:"max_score"`   //最高分
	StudentNum uint    `json:"student_num"` //学习人数
}

// CourseService 课程模块服务接口
type CourseService interface {
	GetCourses(keyword string, order string, page uint) ([]database.Course, error)              // 获取课程列表
	GetCourseAPI(id uint, uid uint) (CourseAPI, error)                                          // 获取课程信息
	GetCourseHistory(courseNumber string) ([]CourseHistoryAPI, error)                           // 获取课程历史信息
	GetUploadUrl(courseNumber, fileName, typ string, uid uint) (url string, id uint, err error) // 获取课程资料上传链接
	LogUpload(id uint, msg string, uid uint) error                                              // 记录课程资料上传日志
	GetCourseSchedule(cookie, term string) (url, note string, err error)                        // 获取课程表
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
	UpdateUser UserAPI `json:"update_user"`
	Like       bool    `json:"like"`
	Own        bool    `json:"own"`
}

// PaperService 文章模块服务接口
type PaperService interface {
	Get(id, uid uint, admin bool) (PaperInfo, error)                                                                      // 获取文章信息
	Create(title, intro, content string, anonymous, publicEdit bool, uid uint) (uint, error)                              // 创建文章
	Edit(id uint, title, intro, content string, anonymous, publicEdit bool, lastTime float64, uid uint, admin bool) error // 编辑文章
	GetList(keyword string, order string, page uint) ([]PaperAPI, error)                                                  // 获取文章列表
	Delete(id, uid uint, admin bool) error                                                                                // 删除文章
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
	User   UserAPI        `json:"user"`
	Images []ImageAPI     `json:"images"`
	Tags   []string       `json:"tags"`
	Claim  database.Claim `json:"claim"`
	Like   bool           `json:"like"`
	Own    bool           `json:"own"`
}

// PosterService 帖子模块服务接口
type PosterService interface {
	Get(id, uid uint, admin bool) (PosterInfo, error)                                                                                                            // 获取帖子信息
	Create(title, text string, imageMids []string, plugins string, anonymous bool, tags []string, claimID uint, public bool, uid uint, admin bool) (uint, error) // 创建帖子
	Edit(id uint, title, text string, imageMids []string, plugins string, anonymous bool, tags []string, claimID uint, public bool, uid uint, admin bool) error  // 编辑帖子
	GetList(mode string, page uint, keyword, order string, uid uint, noAnonymous bool, onlyPublic bool) ([]PosterAPI, error)                                     // 获取帖子列表
	Delete(id, uid uint, admin bool) error                                                                                                                       // 删除帖子
	GetClaims() []database.Claim                                                                                                                                 // 获取声明列表
}
