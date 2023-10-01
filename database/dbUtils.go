/**
* @author:YHCnb
* Package:
* @date:2023/9/30 21:52
* Description:
 */
package database

import (
	"errors"
	"gorm.io/gorm"
	"reflect"
	"time"
)

// StructToMap 将结构体转换为map
func StructToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		// 获取字段的JSON标签
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" {
			tag = fieldName
		}
		// 如果JSON标签为BASE，表示是Base结构体，将其扁平化处理
		if tag == "Base" {
			baseFields := StructToMap(field.Interface())
			for k, v := range baseFields {
				result[k] = v
			}
		} else {
			result[tag] = field.Interface()
		}
	}
	return result
}

// MapToStruct 将map转换为结构体
func MapToStruct(data map[string]interface{}, targetStruct interface{}) error {
	destValue := reflect.ValueOf(targetStruct)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a struct")
	}

	val := destValue.Elem()
	typ := reflect.TypeOf(targetStruct).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		// 获取字段的JSON标签
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" {
			tag = fieldName
		}
		// 如果JSON标签为Base，表示是Base结构体
		if tag == "Base" {
			baseStruct := Base{
				ID:        uint(data["id"].(float64)),
				CreatedAt: data["create_time"].(time.Time),
				UpdatedAt: data["update_time"].(time.Time),
				DeletedAt: data["delete_time"].(gorm.DeletedAt),
			}
			field.Set(reflect.ValueOf(baseStruct))
		} else {
			if value, ok := data[tag]; ok {
				field.Set(reflect.ValueOf(value))
			}
		}
	}
	return nil
}
