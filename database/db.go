/*
 * @Author: flwfdd
 * @Date: 2023-03-20 09:51:48
 * @LastEditTime: 2023-10-10 19:53:47
 * @Description: _(:з」∠)_
 */
package database

import (
	"BIT101-GO/util/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var ClaimMap map[uint]Claim
var IdentityMap map[uint]Identity
var ReportTypeMap map[uint]ReportType
var BanMap map[uint]Ban

// 枚举用户类型(需要与数据库中定义一致)
const (
	Identity_Super = iota
	Identity_Normal
	Identity_Admin
	Identity_Organization
	Identity_Robot
)

// 基本模型
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"create_time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"delete_time"`
}

// 用户
type User struct {
	Base
	Sid      string `gorm:"not null;uniqueIndex" json:"sid"`
	Password string `gorm:"not null" json:"password"`
	Nickname string `gorm:"not null;unique" json:"nickname"`
	Avatar   string `json:"avatar"`
	Motto    string `json:"motto"`
	Identity uint   `gorm:"default:1" json:"identity"`
}

// 身份
type Identity struct {
	Base
	Text  string `gorm:"not null" json:"text"`
	Color string `json:"color"`
}

// 图片
type Image struct {
	Base
	Mid  string `gorm:"not null;uniqueIndex" json:"mid"`
	Size uint   `gorm:"not null" json:"size"`
	Uid  uint   `gorm:"not null" json:"uid"`
}

// 标签
type Tag struct {
	Base
	Name string `gorm:"not null;unique" json:"name"` //标签名
	Hot  uint   `gorm:"default:0" json:"hot"`        //热度
}

// 申明
type Claim struct {
	Base
	Text string `gorm:"not null;unique" json:"text"` //申明内容
}

// 帖子
type Poster struct {
	Base
	Title      string    `gorm:"not null" json:"title"`           //标题
	Text       string    `gorm:"not null" json:"text"`            //内容
	Images     string    `json:"images"`                          //图片mids，以" "拼接
	Uid        uint      `gorm:"not null;index" json:"uid"`       //用户id
	Anonymous  bool      `json:"anonymous"`                       //是否匿名
	Public     bool      `json:"public"`                          //是否可见
	LikeNum    uint      `gorm:"default:0" json:"like_num"`       //点赞数
	CommentNum uint      `gorm:"default:0" json:"comment_num"`    //评论数
	Tags       string    `json:"tags"`                            //标签，以" "拼接
	ClaimID    uint      `json:"claim_id"`                        //申明id
	Plugins    string    `json:"plugins"`                         //插件
	EditAt     time.Time `gorm:"autoCreateTime" json:"edit_time"` //最后编辑时间
}

// 文章
type Paper struct {
	Base
	Title      string    `gorm:"not null" json:"title"`           //标题
	Intro      string    `json:"intro"`                           //简介
	Content    string    `json:"content"`                         //内容
	CreateUid  uint      `gorm:"not null" json:"create_uid"`      //最初发布用户id
	UpdateUid  uint      `gorm:"not null" json:"update_uid"`      //最后编辑用户id
	Anonymous  bool      `gorm:"default:false" json:"anonymous"`  //是否匿名
	LikeNum    uint      `gorm:"default:0" json:"like_num"`       //点赞数
	CommentNum uint      `gorm:"default:0" json:"comment_num"`    //评论数
	PublicEdit bool      `gorm:"default:true" json:"public_edit"` //是否共享编辑
	EditAt     time.Time `gorm:"autoCreateTime" json:"edit_time"` //最后编辑时间
}

// 文章记录
type PaperHistory struct {
	Base
	Pid       uint   `gorm:"not null" json:"pid"`            //文章id
	Title     string `gorm:"not null" json:"title"`          //标题
	Intro     string `json:"intro"`                          //简介
	Content   string `json:"content"`                        //内容
	Uid       uint   `gorm:"not null" json:"uid"`            //用户id
	Anonymous bool   `gorm:"default:false" json:"anonymous"` //是否匿名
}

// 点赞
type Like struct {
	Base
	Obj string `gorm:"not null;index" json:"obj"` //点赞对象
	Uid uint   `gorm:"not null;index" json:"uid"` //用户id
}

// 关注
type Follow struct {
	Base
	Uid       uint `gorm:"not null;index" json:"uid"`        //用户id
	FollowUid uint `gorm:"not null;index" json:"follow_uid"` //关注用户id
}

// 评论
type Comment struct {
	Base
	Obj        string `gorm:"not null;index" json:"obj"`      //评论对象
	Uid        uint   `gorm:"not null;index" json:"uid"`      //用户id
	Text       string `gorm:"not null" json:"text"`           //评论内容
	Anonymous  bool   `gorm:"default:false" json:"anonymous"` //是否匿名
	LikeNum    uint   `gorm:"default:0" json:"like_num"`      //点赞数
	CommentNum uint   `gorm:"default:0" json:"comment_num"`   //评论数
	ReplyObj   string `json:"reply_obj"`                      //回复对象
	ReplyUid   int    `gorm:"default:0" json:"reply_uid"`     //回复用户id
	Rate       uint   `gorm:"default:0" json:"rate"`          //评分
	Images     string `json:"images"`                         //图片mids，以" "拼接
}

// 课程
type Course struct {
	Base
	Name           string    `gorm:"not null" json:"name"`                      //课程名
	Number         string    `gorm:"not null;index" json:"number"`              //课程号
	LikeNum        uint      `gorm:"default:0" json:"like_num"`                 //点赞数
	CommentNum     uint      `gorm:"default:0" json:"comment_num"`              //评论数
	Rate           float64   `gorm:"default:0" json:"rate"`                     //评分
	RateSum        uint      `gorm:"default:0" json:"rate_sum"`                 //评分总和
	TeachersName   string    `gorm:"not null" json:"teachers_name"`             //教师名
	TeachersNumber string    `gorm:"not null" json:"teachers_number"`           //教师号
	Teachers       []Teacher `gorm:"many2many:course_teachers" json:"teachers"` //教师
}

// 教师
type Teacher struct {
	Base
	Name   string `gorm:"not null" json:"name"`   //教师名
	Number string `gorm:"not null" json:"number"` //教师号
}

// 课程历史
type CourseHistory struct {
	Base
	Number     string  `gorm:"not null;index" json:"number"` //课程号
	Term       string  `gorm:"not null;index" json:"term"`   //学期
	AvgScore   float64 `gorm:"default:0" json:"avg_score"`   //均分
	MaxScore   float64 `gorm:"default:0" json:"max_score"`   //最高分
	StudentNum uint    `gorm:"default:0" json:"student_num"` //学习人数
}

// 课程资料上传记录
type CourseUploadLog struct {
	Base
	Uid          uint   `gorm:"not null;index" json:"uid"`           //用户id
	CourseNumber string `gorm:"not null;index" json:"course_number"` //课程号
	CourseName   string `gorm:"not null" json:"course_name"`         //课程名
	Type         string `gorm:"not null" json:"type"`                //资料类型
	Name         string `gorm:"not null" json:"name"`                //资料名
	Msg          string `json:"msg"`                                 //备注
	Finish       bool   `gorm:"default:false" json:"finish"`         //是否完成
}

// 课程资料文档
type CourseUploadReadme struct {
	Base
	CourseNumber string `gorm:"not null;index" json:"course_number"` //课程号
	Text         string `json:"text"`                                //内容
}

// 变量
type Variable struct {
	Base
	Obj  string `gorm:"not null;index" json:"obj"` //对象
	Data string `json:"data"`
}

// 消息摘要
type MessageSummary struct {
	Base
	Uid       uint      `gorm:"not null;index" json:"uid"`       //用户id
	Type      string    `gorm:"not null;index" json:"type"`      //对象
	UnreadNum uint      `gorm:"default:0" json:"unread_num"`     //未读数
	LastTime  time.Time `gorm:"autoCreateTime" json:"last_time"` //最后时间
	Content   string    `gorm:"not null" json:"content"`         //内容
}

// 消息
type Message struct {
	Base
	Obj     string `gorm:"not null;index" json:"obj"`      //对象
	FromUid int    `gorm:"not null;index" json:"from_uid"` //发起用户id
	ToUid   uint   `gorm:"not null;index" json:"to_uid"`   //接收用户id
	Type    string `json:"type"`                           //类型
	LinkObj string `json:"link_obj"`                       //跳转对象
	Content string `json:"content"`                        //内容
}

// 举报
type Report struct {
	Base
	Obj    string `gorm:"not null;index" json:"obj"`     //举报对象
	Text   string `gorm:"not null" json:"text"`          //举报理由
	Uid    uint   `gorm:"not null;index" json:"uid"`     //举报人id
	Status int    `gorm:"default:0" json:"status"`       //状态 0为未处理 1为举报成功 2为举报失败
	TypeId uint   `gorm:"not null;index" json:"type_id"` //举报类型id
}

// 举报类型
type ReportType struct {
	Base
	Text string `gorm:"not null" json:"text"` //类型内容
}

// 小黑屋
type Ban struct {
	Base
	Uid  uint   `gorm:"not null;unique" json:"uid"` //封禁id
	Time string `gorm:"not null" json:"time"`       //解封时间
}

// InitMaps 初始化Map(针对常用且稳定的数据)
func InitMaps() {
	//初始化 ClaimMap
	ClaimMap = make(map[uint]Claim)
	var claims []Claim
	DB.Find(&claims)
	for _, claim := range claims {
		ClaimMap[claim.ID] = claim
	}
	//初始化 IdentityMap
	IdentityMap = make(map[uint]Identity)
	var identities []Identity
	DB.Find(&identities)
	for _, identity := range identities {
		IdentityMap[identity.ID] = identity
	}
	//初始化 ReportTypeMap
	ReportTypeMap = make(map[uint]ReportType)
	var reportTypes []ReportType
	DB.Find(&reportTypes)
	for _, reportType := range reportTypes {
		ReportTypeMap[reportType.ID] = reportType
	}
	// 初始化 BanMap
	BanMap = make(map[uint]Ban)
	var bans []Ban
	DB.Find(&bans)
	for _, ban := range bans {
		BanMap[ban.Uid] = ban
	}
}

func Init() {
	dsn := config.Config.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	db.AutoMigrate(
		&User{}, &Image{}, &Paper{}, &PaperHistory{}, &Like{}, &Comment{}, &Course{}, &CourseHistory{},
		&Teacher{}, &CourseUploadLog{}, &CourseUploadReadme{}, &Variable{}, &Message{}, &MessageSummary{},
		&Tag{}, &Claim{}, &Poster{}, &Identity{}, &Follow{}, &Report{}, &ReportType{}, &Ban{},
	)
	InitMaps()
}
