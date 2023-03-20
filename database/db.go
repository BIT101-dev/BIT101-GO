/*
 * @Author: flwfdd
 * @Date: 2023-03-20 09:51:48
 * @LastEditTime: 2023-03-20 14:42:40
 * @Description: _(:з」∠)_
 */
package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	gorm.Model
	Sid      string `gorm:"not null;uniqueIndex"`
	Password string `gorm:"not null"`
	Nickname string `gorm:"not null;unique"`
	Avatar   string
	Motto    string
	Level    int `gorm:"default:1"`
}

func Init() {
	dsn := "host=localhost user=bit101 password=BIT101 dbname=bit101 port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db

	db.AutoMigrate(&User{})
}
