/*
 * @Author: flwfdd
 * @Date: 2023-03-20 16:41:23
 * @LastEditTime: 2023-03-21 01:24:35
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/saver"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// 将图片mid转换为url
func GetImageUrl(mid string) string {
	if mid == "" {
		return saver.GetUrl(filepath.Join("img", config.Config.DefaultAvatar))
	}
	return saver.GetUrl(filepath.Join("img", mid))
}

// 通过文件上传图片
func ImageUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 获取图片二进制串
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"msg": "读取上传文件出错Orz"})
		return
	}
	defer src.Close()

	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, src); err != nil {
		c.JSON(500, gin.H{"msg": "读取上传文件出错Orz"})
		return
	}

	// 获取图片扩展名
	ext := getImageExt(buff.Bytes())
	if ext == "" {
		c.JSON(500, gin.H{"msg": "不支持该图片类型Orz"})
		return
	}

	// 计算文件md5 生成文件名
	h := md5.New()
	h.Write(buff.Bytes())
	mid := hex.EncodeToString(h.Sum(nil)) + ext
	path := filepath.Join("img", mid)

	image := database.Image{}
	database.DB.Limit(1).Find(&image, "mid = ?", mid)
	if image.Mid != "" {
		c.JSON(200, gin.H{"url": saver.GetUrl(path), "mid": image.Mid})
		return
	}

	// 保存文件
	err = saver.Save(path, buff.Bytes())
	if err != nil {
		c.JSON(500, gin.H{"msg": "保存文件出错Orz"})
		return
	}

	image = database.Image{
		Mid:  mid,
		Size: uint(file.Size),
		User: c.GetUint("uid"),
	}
	database.DB.Create(&image)

	c.JSON(200, gin.H{"url": saver.GetUrl(path), "mid": image.Mid})
}

// 获取图片后缀名 无效返回空字符串
func getImageExt(content []byte) string {
	switch http.DetectContentType(content) {
	case "image/jpg":
		return ".jpg"
	case "image/jpeg":
		return ".jpeg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
