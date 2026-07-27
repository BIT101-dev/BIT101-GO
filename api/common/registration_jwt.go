/*
 * @Author: yht0511
 * @Date: 2026-07-27 16:45:05
 * @LastEditTime: 2026-07-27 16:49:01
 * @Description: 统一身份验证JWT验证模块
 */
package common

import (
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type registrationJWTClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

// 校验JWT合法性
func VerifyRegistrationJWT(tokenString, publicKeyFile, issuer, audience string) (string, error) {
	if strings.TrimSpace(tokenString) == "" {
		return "", errors.New("注册JWT为空Orz")
	}
	if publicKeyFile == "" || issuer == "" || audience == "" {
		return "", errors.New("注册JWT验证未配置Orz")
	}

	publicKeyPEM, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return "", errors.New("读取注册JWT公钥失败Orz")
	}
	publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", errors.New("解析注册JWT公钥失败Orz")
	}

	claims := &registrationJWTClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodEdDSA {
				return nil, errors.New("注册JWT签名算法无效Orz")
			}
			return publicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil || !token.Valid {
		return "", errors.New("注册JWT无效Orz")
	}
	if claims.Purpose != "registration" {
		return "", errors.New("注册JWT用途无效Orz")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return "", errors.New("注册JWT缺少学号Orz")
	}
	return claims.Subject, nil
}
