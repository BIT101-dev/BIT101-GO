/*
 * @Author: flwfdd
 * @Date: 2023-03-25 15:23:50
 * @LastEditTime: 2023-03-25 15:34:30
 * @Description: _(:з」∠)_
 */

/*
	优化：

错误响应复用: 抽取 respondWithError 方法，减少重复代码，提高可读性。
状态码调整:
参数错误返回 400 Bad Request。
数据不存在返回 404 Not Found。
数据库错误返回 500 Internal Server Error。
查询优化: 将 Find 替换为 First，提高效率。
幂等性支持: VariablePost 支持创建新变量（如果记录不存在）。
响应信息细化: 更精确地描述操作结果和错误原因。
通用性增强: 更新和插入逻辑集成到一起，支持更灵活的数据库操作。
*/
package controller

import (
	"BIT101-GO/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取变量请求结构
type VariableGetRequest struct {
	Obj string `form:"obj" binding:"required"`
}

// 设置变量请求结构
type VariablePostRequest struct {
	Obj  string `json:"obj" binding:"required"`
	Data string `json:"data" binding:"required"`
}

// 统一响应错误
func respondWithError(c *gin.Context, statusCode int, msg string) {
	c.JSON(statusCode, gin.H{"msg": msg})
}

// 获取变量
func VariableGet(c *gin.Context) {
	var query VariableGetRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		respondWithError(c, http.StatusBadRequest, "参数错误awa")
		return
	}

	var variable database.Variable
	// 查询第一条记录
	if err := database.DB.Where("obj = ?", query.Obj).First(&variable).Error; err != nil {
		if err.Error() == "record not found" {
			respondWithError(c, http.StatusNotFound, "变量不存在Orz")
		} else {
			respondWithError(c, http.StatusInternalServerError, "数据库错误Orz")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variable.Data})
}

// 设置变量
func VariablePost(c *gin.Context) {
	var query VariablePostRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		respondWithError(c, http.StatusBadRequest, "参数错误awa")
		return
	}

	// 使用 `Save` 方法支持更新或插入
	var variable database.Variable
	if err := database.DB.Where("obj = ?", query.Obj).First(&variable).Error; err != nil {
		if err.Error() == "record not found" {
			// 如果记录不存在，插入新数据
			variable = database.Variable{Obj: query.Obj, Data: query.Data}
			if err := database.DB.Create(&variable).Error; err != nil {
				respondWithError(c, http.StatusInternalServerError, "无法插入新变量Orz")
				return
			}
			c.JSON(http.StatusOK, gin.H{"msg": "变量创建成功OvO"})
			return
		} else {
			respondWithError(c, http.StatusInternalServerError, "数据库错误Orz")
			return
		}
	}

	// 记录存在，更新数据
	if err := database.DB.Model(&variable).Update("data", query.Data).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "更新失败Orz")
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功OvO"})
}
