/*
 * @Author: flwfdd
 * @Date: 2023-03-20 16:41:23
 * @LastEditTime: 2023-03-29 20:23:30
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/saver"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ImageAPI struct {
	Mid    string `json:"mid"`
	Url    string `json:"url"`
	LowURL string `json:"low_url"`
}

// GetImageAPI 生成ImageAPI
func GetImageAPI(mid string) ImageAPI {
	return ImageAPI{
		Mid:    mid,
		Url:    GetImageUrl(mid),
		LowURL: GetImageUrl(mid) + config.GetConfig().Saver.ImageUrlSuffix,
	}
}

// GetImageAPIArr 生成ImageAPIArr
func GetImageAPIArr(mids []string) []ImageAPI {
	imageAPIArr := []ImageAPI{}
	for i := range mids {
		imageAPIArr = append(imageAPIArr, GetImageAPI(mids[i]))
	}
	return imageAPIArr
}

// GetImageUrl 将图片mid转换为url
func GetImageUrl(mid string) string {
	if mid == "" {
		return saver.GetUrl(filepath.Join("img", config.GetConfig().DefaultAvatar))
	}
	return saver.GetUrl(filepath.Join("img", mid))
}

// CheckImage 检验mids是否有效
func CheckImage(mids []string) bool {
	for i := range mids {
		if mids[i] == "" {
			continue
		}
		image := database.Image{}
		database.DB.Limit(1).Find(&image, "mid = ?", mids[i])
		if image.Mid == "" {
			return false
		}
	}
	return true
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

// 保存文件
func save(c *gin.Context, content []byte) {
	// 获取图片扩展名
	ext := getImageExt(content)
	if ext == "" {
		c.JSON(500, gin.H{"msg": "不支持的图片类型Orz"})
		return
	}

	// 计算文件md5 生成文件名
	h := md5.New()
	h.Write(content)
	mid := hex.EncodeToString(h.Sum(nil)) + ext
	path := filepath.Join("img", mid)

	image := database.Image{}
	database.DB.Limit(1).Find(&image, "mid = ?", mid)
	if image.Mid != "" {
		c.JSON(200, GetImageAPI(mid))
		return
	}

	// 保存文件
	_, err := saver.Save(path, content)
	if err != nil {
		c.JSON(500, gin.H{"msg": "保存图片出错Orz"})
		return
	}

	image = database.Image{
		Mid:  mid,
		Size: uint(len(content)),
		Uid:  c.GetUint("uid_uint"),
	}
	database.DB.Create(&image)

	c.JSON(200, GetImageAPI(mid))
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
		c.JSON(500, gin.H{"msg": "读取上传图片出错Orz"})
		return
	}
	defer src.Close()

	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, src); err != nil {
		c.JSON(500, gin.H{"msg": "读取上传图片出错Orz"})
		return
	}

	save(c, buff.Bytes())
}

// 通过url上传图片请求结构
type ImageUploadByUrlRequest struct {
	Url string `json:"url" binding:"required"`
}

// 通过url上传图片
func ImageUploadByUrl(c *gin.Context) {
	var query ImageUploadByUrlRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	req, _ := http.NewRequest("GET", query.Url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"msg": "获取图片出错Orz"})
		return
	}
	defer resp.Body.Close()
	reader := http.MaxBytesReader(nil, resp.Body, config.GetConfig().Saver.MaxSize<<20) //限制请求大小
	body, err := io.ReadAll(reader)
	if err != nil {
		c.JSON(500, gin.H{"msg": "图片过大Orz"})
		return
	}

	save(c, body)
}
