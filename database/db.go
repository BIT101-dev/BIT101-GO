/*
 * @Author: flwfdd
 * @Date: 2023-03-20 09:51:48
 * @LastEditTime: 2023-03-21 01:20:14
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
	DeletedAt gorm.DeletedAt `gorm:"index"`
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
	User uint   `gorm:"not null" json:"user"`
}

func Init() {
	dsn := config.Config.Dsn
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	db.AutoMigrate(&User{}, &Image{})
}
