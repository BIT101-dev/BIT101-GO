/*
 * @Author: flwfdd
 * @Date: 2025-03-09 17:42:45
 * @LastEditTime: 2025-03-11 00:19:45
 * @Description: _(:з」∠)_
 */
package types

import (
	"BIT101-GO/database"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ObjHandler interface {
	IsExist(id uint) bool
	GetObjType() string
	GetUid(id uint) (uint, error)
	LikeHandler(tx *gorm.DB, id uint, delta int, uid uint) (likeNum uint, err error)
	CommentHandler(tx *gorm.DB, id uint, comment database.Comment, delta int, uid uint) (commentNum uint, err error)
}

type Obj struct {
	id      uint
	handler ObjHandler
}

var objHandlers = make(map[string]ObjHandler)

func RegisterObjHandler(h ObjHandler) {
	objHandlers[h.GetObjType()] = h
}

func NewObj(obj string) (Obj, error) {
	for k, v := range objHandlers {
		if strings.HasPrefix(obj, k) {
			id, err := strconv.ParseUint(obj[len(k):], 10, 32)
			if err != nil {
				return Obj{}, err
			}
			if !v.IsExist(uint(id)) {
				return Obj{}, errors.New("对象不存在Orz")
			}
			return Obj{
				id:      uint(id),
				handler: v,
			}, nil
		}
	}
	return Obj{}, errors.New("未知对象类型Orz")
}

func (o Obj) GetID() uint {
	return o.id
}

func (o Obj) GetObjID() string {
	if o.id == 0 {
		return ""
	}
	return fmt.Sprintf("%s%d", o.handler.GetObjType(), o.id)
}

func (o Obj) GetObjType() string {
	return o.handler.GetObjType()
}

func (o Obj) GetUid() (uint, error) {
	return o.handler.GetUid(o.id)
}

func (o Obj) LikeHandler(tx *gorm.DB, delta int, uid uint) (likeNum uint, err error) {
	return o.handler.LikeHandler(tx, o.id, delta, uid)
}

func (o Obj) CommentHandler(tx *gorm.DB, comment database.Comment, delta int, uid uint) (commentNum uint, err error) {
	return o.handler.CommentHandler(tx, o.id, comment, delta, uid)
}
