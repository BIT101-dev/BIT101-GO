package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/database"
	"errors"
)

// 检查实现了VariableService接口
var _ types.VariableService = (*VariableService)(nil)

// VariableService 变量模块服务
type VariableService struct{}

// NewVariableService 创建变量模块服务
func NewVariableService() *VariableService {
	return &VariableService{}
}

// Get 获取变量
func (s *VariableService) Get(objID string) (string, error) {
	var variable database.Variable
	if err := database.DB.Where("obj = ?", objID).Limit(1).Find(&variable).Error; err != nil {
		return "", errors.New("数据库错误Orz")
	}
	if variable.ID == 0 {
		return "", errors.New("变量不存在Orz")
	}
	return variable.Data, nil
}

// Set 设置变量
func (s *VariableService) Set(objID, data string) error {
	if err := database.DB.Model(&database.Variable{}).Where("obj = ?", objID).Update("data", data).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	return nil
}
