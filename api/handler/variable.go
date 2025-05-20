/*
 * @Author: flwfdd
 * @Date: 2023-03-25 15:23:50
 * @LastEditTime: 2025-03-11 01:46:51
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/service"

	"github.com/gin-gonic/gin"
)

// 变量模块接口
type VariableHandler struct {
	VariableService *service.VariableService
}

// 创建变量模块接口
func NewVariableHandler(variableService *service.VariableService) *VariableHandler {
	return &VariableHandler{variableService}
}

// 获取变量
func (h *VariableHandler) GetHandler(c *gin.Context) {
	type Request struct {
		Obj string `form:"obj" binding:"required"`
	}

	type Response struct {
		Data string `json:"data"`
	}

	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	data, err := h.VariableService.Get(query.Obj)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{data})
}

// 设置变量
func (h *VariableHandler) SetHandler(c *gin.Context) {
	type Request struct {
		Obj  string `json:"obj" binding:"required"`
		Data string `json:"data" binding:"required"`
	}

	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	if common.HandleError(c, h.VariableService.Set(query.Obj, query.Data)) {
		return
	}

	c.JSON(200, gin.H{"msg": "更新成功OvO"})
}
