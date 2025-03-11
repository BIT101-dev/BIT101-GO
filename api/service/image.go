/*
 * @Author: flwfdd
 * @Date: 2023-03-20 16:41:23
 * @LastEditTime: 2025-03-09 13:32:45
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/saver"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"path/filepath"
)

// 检查实现了ImageService接口
var _ types.ImageService = (*ImageService)(nil)

// ImageService 图片模块服务
type ImageService struct{}

// NewImageService 创建图片模块服务
func NewImageService() *ImageService {
	return &ImageService{}
}

// GetImageAPI 通过mid生成ImageAPI
func (s *ImageService) GetImageAPI(mid string) types.ImageAPI {
	return types.ImageAPI{
		Mid:    mid,
		Url:    s.Mid2Url(mid),
		LowURL: s.Mid2Url(mid) + config.Get().Saver.ImageUrlSuffix,
	}
}

// GetImageAPIList 通过mids生成ImageAPI数组
func (s *ImageService) GetImageAPIList(mids []string) []types.ImageAPI {
	imageAPIArr := []types.ImageAPI{}
	for i := range mids {
		imageAPIArr = append(imageAPIArr, s.GetImageAPI(mids[i]))
	}
	return imageAPIArr
}

// Mid2Url 将图片mid转换为url
func (s *ImageService) Mid2Url(mid string) string {
	if mid == "" {
		// 返回默认头像
		return saver.GetUrl(filepath.Join("img", config.Get().DefaultAvatar))
	}
	return saver.GetUrl(filepath.Join("img", mid))
}

// CheckMid 检查图片mid是否存在
func (s *ImageService) CheckMid(mid string) bool {
	return database.DB.First(&database.Image{}, "mid = ?", mid).Error == nil
}

// CheckMids 批量检查图片mid是否存在
func (s *ImageService) CheckMids(mids []string) bool {
	if len(mids) == 0 {
		return true
	}

	var count int64
	database.DB.Model(&database.Image{}).Where("mid IN (?)", mids).Count(&count)
	return count == int64(len(mids))
}

// getImageExt 获取图片后缀名 无效返回空字符串
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

// save 保存图片
func (s *ImageService) Save(uid uint, content []byte) (types.ImageAPI, error) {
	// 获取图片扩展名
	ext := getImageExt(content)
	if ext == "" {
		return types.ImageAPI{}, errors.New("不支持的图片类型Orz")
	}

	// 计算文件md5 生成文件名
	h := md5.New()
	h.Write(content)
	mid := hex.EncodeToString(h.Sum(nil)) + ext
	path := filepath.Join("img", mid)

	if s.CheckMid(mid) {
		return s.GetImageAPI(mid), nil
	}

	// 保存文件
	_, err := saver.Save(path, content)
	if err != nil {
		return types.ImageAPI{}, errors.New("保存图片出错Orz")
	}

	image := database.Image{
		Mid:  mid,
		Size: uint(len(content)),
		Uid:  uid,
	}
	database.DB.Create(&image)

	return s.GetImageAPI(mid), nil
}
