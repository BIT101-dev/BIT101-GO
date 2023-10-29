/*
 * @Author: flwfdd
 * @Date: 2023-03-25 15:23:50
 * @LastEditTime: 2023-03-25 15:34:30
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"

	"github.com/gin-gonic/gin"
)

// 获取变量请求结构
type VariableGetRequest struct {
	Obj string `form:"obj" binding:"required"`
}

// 获取变量
func VariableGet(c *gin.Context) {
	var query VariableGetRequest
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var variable database.Variable
	if err := database.DB.Where("obj = ?", query.Obj).Limit(1).Find(&variable).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if variable.ID == 0 {
		c.JSON(500, gin.H{"msg": "变量不存在Orz"})
		return
	}
	c.JSON(200, gin.H{"data": variable.Data})
}

// 设置变量请求结构
type VariablePostRequest struct {
	Obj  string `json:"obj" binding:"required"`
	Data string `json:"data" binding:"required"`
}

// 设置变量
func VariablePost(c *gin.Context) {
	if !c.GetBool("super") {
		c.JSON(401, gin.H{"msg": "没有权限Orz"})
		return
	}

	var query VariablePostRequest
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	if err := database.DB.Model(&database.Variable{}).Where("obj = ?", query.Obj).Update("data", query.Data).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "更新成功OvO"})
}
