/*
 * @Author: flwfdd
 * @Date: 2023-03-20 09:51:48
 * @LastEditTime: 2023-03-21 20:58:52
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

type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"create_time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"delete_time"`
}

type User struct {
	Base
	Sid      string `gorm:"not null;uniqueIndex" json:"sid"`
	Password string `gorm:"not null" json:"password"`
	Nickname string `gorm:"not null;unique" json:"nickname"`
	Avatar   string `json:"avatar"`
	Motto    string `json:"motto"`
	Level    int    `gorm:"default:1" json:"level"`
}

type Image struct {
	Base
	Mid  string `gorm:"not null;uniqueIndex" json:"mid"`
	Size uint   `gorm:"not null" json:"size"`
	Uid  uint   `gorm:"not null" json:"uid"`
}

type Paper struct {
	Base
	Title      string `gorm:"not null" json:"title"`           //标题
	Intro      string `json:"intro"`                           //简介
	Content    string `json:"content"`                         //内容
	CreateUid  uint   `gorm:"not null" json:"create_uid"`      //最初发布用户id
	UpdateUid  uint   `gorm:"not null" json:"update_uid"`      //最后编辑用户id
	Anonymous  bool   `gorm:"default:false" json:"anonymous"`  //是否匿名
	LikeNum    uint   `gorm:"default:0" json:"like_num"`       //点赞数
	CommentNum uint   `gorm:"default:0" json:"comment_num"`    //评论数
	PublicEdit bool   `gorm:"default:true" json:"public_edit"` //是否共享编辑
}

type PaperHistory struct {
	Base
	Pid       uint   `gorm:"not null" json:"pid"`            //文章id
	Title     string `gorm:"not null" json:"title"`          //标题
	Intro     string `json:"intro"`                          //简介
	Content   string `json:"content"`                        //内容
	Uid       uint   `gorm:"not null" json:"uid"`            //用户id
	Anonymous bool   `gorm:"default:false" json:"anonymous"` //是否匿名
}

func Init() {
	dsn := config.Config.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	db.AutoMigrate(&User{}, &Image{}, &Paper{}, &PaperHistory{})
}
