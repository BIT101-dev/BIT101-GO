/*
 * @Author: flwfdd
 * @Date: 2023-03-19 23:29:07
 * @LastEditTime: 2023-03-20 02:11:21
 * @Description: _(:з」∠)_
 */
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 用于验证用户登录状态
type UserClaims struct {
	Id     string `json:"id"`
	Expire int64  `json:"exp"`
	jwt.RegisteredClaims
}

// 生成用户token 也可将验证码作为key来验证
func GetUserToken(id string, expire int64, key string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		id,
		time.Now().Unix() + expire,
		jwt.RegisteredClaims{},
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		panic(err)
	}
	return tokenString
}

// 验证用户 返回用户id和是否成功
func VeirifyUserToken(tokenString string, key string) (string, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return "", false
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid && claims.Expire > time.Now().Unix() {
		return claims.Id, true
	} else {
		return "", false
	}
}
