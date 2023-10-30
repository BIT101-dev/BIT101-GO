/**
* @author:YHCnb
* Package:
* @date:2023/10/28 21:27
* Description:
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// ReportTypeListGet 获取举报类型列表
func ReportTypeListGet(c *gin.Context) {
	reportTypes := make([]database.ReportType, 0)
	for _, reportType := range database.ReportTypeMap {
		reportTypes = append(reportTypes, reportType)
	}
	c.JSON(200, reportTypes)
}

// ReportPostQuery 举报请求接口
type ReportPostQuery struct {
	Obj    string `json:"obj"`
	TypeId uint   `json:"type_id"`
	Text   string `json:"text"`
}

// ReportPost 举报
func ReportPost(c *gin.Context) {
	var query ReportPostQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(500, gin.H{"msg": "参数错误Orz"})
		return
	}
	if _, ok := database.ReportTypeMap[query.TypeId]; !ok {
		c.JSON(500, gin.H{"msg": "举报类型不存在Orz"})
		return
	}
	obj_type, obj_id := getTypeID(query.Obj)
	var isExist bool
	switch obj_type {
	case "user":
		var user database.User
		database.DB.Where("id = ?", obj_id).Limit(1).Find(&user)
		if user.ID != 0 {
			isExist = true
		}
	case "poster":
		var poster database.Poster
		database.DB.Where("id = ?", obj_id).Limit(1).Find(&poster)
		if poster.ID != 0 {
			isExist = true
		}
	case "comment":
		var comment database.Comment
		database.DB.Where("id = ?", obj_id).Limit(1).Find(&comment)
		if comment.ID != 0 {
			isExist = true
		}
	default:
		isExist = false
	}
	if !isExist {
		c.JSON(500, gin.H{"msg": "举报对象不存在Orz"})
		return
	}
	report := database.Report{
		Obj:    query.Obj,
		Text:   query.Text,
		Uid:    c.GetUint("uid_uint"),
		TypeId: query.TypeId,
	}
	database.DB.Create(&report)
	c.JSON(200, gin.H{"msg": "举报成功"})
}

// ReportPutQuery 修改举报状态请求接口
type ReportPutQuery struct {
	Status int `form:"status"`
}

// ReportPut 修改举报状态
func ReportPut(c *gin.Context) {
	var query ReportPutQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var report database.Report
	if err := database.DB.Limit(1).Find(&report, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if report.ID == 0 {
		c.JSON(500, gin.H{"msg": "举报不存在Orz"})
		return
	}
	report.Status = query.Status
	database.DB.Save(&report)
	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}

// ReportListQuery 获取举报类型请求接口
type ReportListQuery struct {
	Page   uint   `form:"page"`
	Uid    string `form:"uid"`
	Obj    string `form:"obj"`
	Status int    `form:"status"` // -1为全部，0为未处理 1为举报成功 2为举报失败
}

// ReportListResponseItem 获取举报列表返回结构
type ReportListResponseItem struct {
	CreatedTime string              `json:"created_time"`
	ID          string              `json:"id"`
	Obj         string              `json:"obj"`
	ReportType  database.ReportType `json:"report_type"`
	Status      int                 `json:"status"` // 0为未处理 1为举报成功 2为举报失败
	Text        string              `json:"text"`
	User        UserAPI             `json:"user"`
}

// ReportList 获取举报列表
func ReportList(c *gin.Context) {
	var query ReportListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if !c.GetBool("admin") && !c.GetBool("super") {
		c.JSON(401, gin.H{"msg": "权限不足awa"})
		return
	}
	q := database.DB
	if query.Uid != "" {
		q = q.Where("uid = ?", query.Uid)
	}
	if query.Obj != "" {
		q = q.Where("obj = ?", query.Obj)
	}
	if query.Status != -1 {
		q = q.Where("status = ?", query.Status)
	}
	var reports []database.Report
	if err := q.Order("created_at DESC").Offset(int(query.Page * config.Config.ReportPageSize)).Limit(int(config.Config.ReportPageSize)).Find(&reports).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	response := make([]ReportListResponseItem, 0)
	for _, report := range reports {
		response = append(response, ReportListResponseItem{
			CreatedTime: report.CreatedAt.String(),
			ID:          strconv.Itoa(int(report.ID)),
			Obj:         report.Obj,
			ReportType:  database.ReportTypeMap[report.TypeId],
			Status:      report.Status,
			Text:        report.Text,
			User:        GetUserAPI(int(report.Uid)),
		})
	}
	c.JSON(200, response)
}

// BanPostQuery 关小黑屋请求结构
type BanPostQuery struct {
	Time string `json:"time"` // 解封时间，格式为ISO 8601带时区
	Uid  uint   `json:"uid"`
}

// BanPost 关小黑屋
func BanPost(c *gin.Context) {
	var query BanPostQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if query.Uid == 0 {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if !c.GetBool("admin") && !c.GetBool("super") {
		c.JSON(401, gin.H{"msg": "权限不足awa"})
		return
	}
	_, err := ParseTime(query.Time)
	if err != nil {
		c.JSON(400, gin.H{"msg": "时间格式错误awa"})
		return
	}
	var user database.User
	database.DB.Where("id = ?", query.Uid).Limit(1).Find(&user)
	if user.ID == 0 {
		c.JSON(400, gin.H{"msg": "用户不存在Orz"})
		return
	}
	if user.Identity == database.Identity_Super || user.Identity == database.Identity_Admin {
		c.JSON(400, gin.H{"msg": "此用户无法被封禁Orz"})
		return
	}
	if _, ok := database.BanMap[query.Uid]; ok {
		var _ban database.Ban
		database.DB.Unscoped().Where("uid = ?", query.Uid).Limit(1).Find(&_ban)
		_ban.DeletedAt = gorm.DeletedAt{}
		_ban.Time = query.Time
		go func() { database.DB.Unscoped().Save(_ban) }()
		database.BanMap[query.Uid] = _ban
		c.JSON(200, gin.H{"msg": "修改封禁时间成功OvO"})
		return
	}
	ban := database.Ban{
		Uid:  query.Uid,
		Time: query.Time,
	}
	database.DB.Create(&ban)
	database.BanMap[query.Uid] = ban
	c.JSON(200, gin.H{"msg": "关小黑屋成功OvO"})
}

// BanListQuery 获取小黑屋列表请求结构
type BanListQuery struct {
	Page uint   `form:"page"`
	Uid  string `form:"uid"`
}

// BanListResponseItem 获取小黑屋列表返回结构
type BanListResponseItem struct {
	ID   string  `json:"id"`
	Time string  `json:"time"`
	User UserAPI `json:"user"`
}

// BanList 获取小黑屋列表
func BanList(c *gin.Context) {
	var query BanListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	if !c.GetBool("admin") && !c.GetBool("super") {
		c.JSON(400, gin.H{"msg": "权限不足awa"})
		return
	}
	CheckBans() // 先将过期的封禁删除
	q := database.DB
	if query.Uid != "" {
		q = q.Where("uid = ?", query.Uid)
	}
	bans := make([]database.Ban, 0)
	if err := q.Order("created_at DESC").Offset(int(query.Page * config.Config.BanPageSize)).Limit(int(config.Config.BanPageSize)).Find(&bans).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	response := make([]BanListResponseItem, 0)
	for _, ban := range bans {
		response = append(response, BanListResponseItem{
			ID:   strconv.Itoa(int(ban.ID)),
			Time: ban.Time,
			User: GetUserAPI(int(ban.Uid)),
		})
	}
	c.JSON(200, response)
}

// CheckBans 检查map里的封禁是否过期
func CheckBans() {
	for _, ban := range database.BanMap {
		t, err := ParseTime(ban.Time)
		if err != nil {
			continue
		}
		if GetNowTime().After(t) {
			database.DB.Delete(&ban)
			delete(database.BanMap, ban.Uid)
		}
	}
}

// CheckBan 检查用户是否被封禁
func CheckBan(uid uint) bool {
	if ban, ok := database.BanMap[uid]; ok {
		t, err := ParseTime(ban.Time)
		if err != nil {
			return false
		}
		if GetNowTime().After(t) {
			database.DB.Delete(&ban)
			delete(database.BanMap, ban.Uid)
			return false
		}
		return true
	}
	return false
}
