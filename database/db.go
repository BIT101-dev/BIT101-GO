/*
 * @Author: flwfdd
 * @Date: 2023-03-20 09:51:48
 * @LastEditTime: 2023-03-30 18:00:06
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
	Level    int    `gorm:"default:1" json:"level"`
}

// 图片
type Image struct {
	Base
	Mid  string `gorm:"not null;uniqueIndex" json:"mid"`
	Size uint   `gorm:"not null" json:"size"`
	Uid  uint   `gorm:"not null" json:"uid"`
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
	Tsv        Tsvector  `gorm:"index:,type:gin" json:"tsv"`      //全文搜索
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

// 评论
type Comment struct {
	Base
	Obj        string `gorm:"not null;index" json:"obj"`      //点赞对象
	Uid        uint   `gorm:"not null;index" json:"uid"`      //用户id
	Text       string `gorm:"not null" json:"text"`           //评论内容
	Anonymous  bool   `gorm:"default:false" json:"anonymous"` //是否匿名
	LikeNum    uint   `gorm:"default:0" json:"like_num"`      //点赞数
	CommentNum uint   `gorm:"default:0" json:"comment_num"`   //评论数
	ReplyObj   string `json:"reply_obj"`                      //回复对象
	ReplyUid   int    `gorm:"default:0" json:"reply_uid"`     //回复用户id
	Rate       uint   `gorm:"default:0" json:"rate"`          //评分
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
	Tsv            Tsvector  `gorm:"index:,type:gin" json:"tsv"`                //搜索
}

// 教师
type Teacher struct {
	Base
	Name   string `gorm:"not null" json:"name"`   //教师名
	Number string `gorm:"not null" json:"number"` //教师号
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
	Obj       string    `gorm:"not null;index" json:"obj"`       //对象
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

func Init() {
	dsn := config.Config.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	db.AutoMigrate(&User{}, &Image{}, &Paper{}, &PaperHistory{}, &Like{}, &Comment{}, &Course{}, &Teacher{}, &CourseUploadLog{}, &CourseUploadReadme{}, &Variable{}, &Message{}, &MessageSummary{})
}
