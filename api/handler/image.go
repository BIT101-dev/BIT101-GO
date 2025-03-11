/*
 * @Author: flwfdd
 * @Date: 2025-03-05 13:35:59
 * @LastEditTime: 2025-03-09 23:25:14
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImageHandler 图片模块接口
type ImageHandler struct {
	ImageSvc types.ImageService
}

// NewImageHandler 创建图片模块接口
func NewImageHandler(s types.ImageService) ImageHandler {
	return ImageHandler{s}
}

// ImageUploadHandler 通过文件上传图片
func (h *ImageHandler) UploadHandler(c *gin.Context) {
	type Response struct {
		types.ImageAPI
	}
	file, err := c.FormFile("file")
	if common.HandleErrorWithCode(c, err, 400) {
		return
	}

	// 获取图片二进制串
	src, err := file.Open()
	if common.HandleErrorWithMessage(c, err, "读取上传图片出错Orz") {
		return
	}
	defer src.Close()
	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, src); common.HandleErrorWithMessage(c, err, "读取上传图片出错Orz") {
		return
	}

	imageAPI, err := h.ImageSvc.Save(middleware.MustGetUserContext(c).UIDUint, buff.Bytes())
	if common.HandleErrorWithMessage(c, err, "保存图片出错Orz") {
		return
	}
	c.JSON(200, Response{imageAPI})
}

// UploadByUrlHandler 通过url上传图片
func (h *ImageHandler) UploadByUrlHandler(c *gin.Context) {
	type Request struct {
		Url string `json:"url" binding:"required"`
	}
	type Response struct {
		types.ImageAPI
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	req, _ := http.NewRequest("GET", query.Url, nil)
	resp, err := http.DefaultClient.Do(req)
	if common.HandleErrorWithMessage(c, err, "获取图片出错Orz") {
		return
	}
	defer resp.Body.Close()
	reader := http.MaxBytesReader(nil, resp.Body, config.Get().Saver.MaxSize<<20) //限制请求大小
	body, err := io.ReadAll(reader)
	if common.HandleErrorWithMessage(c, err, "图片过大Orz") {
		return
	}

	imageAPI, err := h.ImageSvc.Save(middleware.MustGetUserContext(c).UIDUint, body)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(200, Response{imageAPI})
}
