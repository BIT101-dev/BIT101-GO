package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 检查实现了ManageService接口
var _ types.ManageService = (*ManageService)(nil)

// ManageService 管理模块服务
type ManageService struct {
	UserSvc types.UserService
}

// NewManageService 创建管理模块服务
func NewManageService(userSvc types.UserService) *ManageService {
	return &ManageService{userSvc}
}

// GetReportTypes 获取举报类型列表
func (s *ManageService) GetReportTypes() []database.ReportType {
	reportTypes := make([]database.ReportType, 0)
	for _, reportType := range database.ReportTypeMap {
		reportTypes = append(reportTypes, reportType)
	}
	return reportTypes
}

// Report 举报
func (s *ManageService) Report(reporter uint, objID string, typeId uint, text string) error {
	if _, ok := database.ReportTypeMap[typeId]; !ok {
		return errors.New("举报类型不存在Orz")
	}
	if _, err := types.NewObj(objID); err != nil {
		return errors.New("举报对象不存在Orz")
	}
	report := database.Report{
		Obj:    objID,
		Text:   text,
		Uid:    reporter,
		TypeId: typeId,
	}
	if err := database.DB.Create(&report).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	return nil
}

// UpdateReportStatus 更新举报状态
func (s *ManageService) UpdateReportStatus(id uint, status int) error {
	var report database.Report
	if err := database.DB.Limit(1).Find(&report, "id = ?", id).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	if report.ID == 0 {
		return errors.New("举报不存在Orz")
	}
	report.Status = status
	if err := database.DB.Save(&report).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	return nil
}

// GetReports 获取举报列表
// 为空或-1时不限制
func (s *ManageService) GetReports(uid int, objID string, status int, page uint) ([]types.ReportAPI, error) {
	q := database.DB
	if uid != -1 {
		q = q.Where("uid = ?", uid)
	}
	if objID != "" {
		q = q.Where("obj = ?", objID)
	}
	if status != -1 {
		q = q.Where("status = ?", status)
	}
	var reports []database.Report
	if err := q.Order("created_at DESC").Offset(int(page * config.Get().ReportPageSize)).Limit(int(config.Get().ReportPageSize)).Find(&reports).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	res, err := FillUsers(
		s.UserSvc,
		reports,
		func(r database.Report) int {
			return int(r.Uid)
		},
		func(r database.Report, u types.UserAPI) types.ReportAPI {
			return types.ReportAPI{
				CreatedTime: r.CreatedAt.String(),
				ID:          int(r.ID),
				Obj:         r.Obj,
				ReportType:  database.ReportTypeMap[r.TypeId],
				Status:      r.Status,
				Text:        r.Text,
				User:        u,
			}
		},
	)
	if err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	return res, nil
}

// Ban 封禁用户
func (s *ManageService) Ban(uid uint, banTime time.Time) error {
	var user database.User
	database.DB.Where("id = ?", uid).Limit(1).Find(&user)
	if user.ID == 0 {
		return errors.New("用户不存在Orz")
	}
	if user.Identity == uint(database.IdentityAdmin) || user.Identity == uint(database.IdentitySuper) {
		return errors.New("此用户无法被封禁Orz")
	}
	if _, ok := database.BanMap[uid]; ok {
		var ban_ database.Ban
		database.DB.Unscoped().Where("uid = ?", uid).Limit(1).Find(&ban_)
		ban_.DeletedAt = gorm.DeletedAt{}
		ban_.Time = banTime
		database.DB.Unscoped().Save(ban_)
		database.BanMap[uid] = ban_
	} else {
		ban := database.Ban{
			Uid:  uid,
			Time: banTime,
		}
		if err := database.DB.Create(&ban).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		database.BanMap[uid] = ban
	}
	return nil
}

// GetBans 获取封禁列表
func (s *ManageService) GetBans(uid int, page uint) ([]types.BanAPI, error) {
	q := database.DB
	if uid != -1 {
		q = q.Where("uid = ?", uid)
	}
	var bans []database.Ban
	if err := q.Order("created_at DESC").Offset(int(page * config.Get().BanPageSize)).Limit(int(config.Get().BanPageSize)).Find(&bans).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	res, err := FillUsers(
		s.UserSvc,
		bans,
		func(b database.Ban) int {
			return int(b.Uid)
		},
		func(b database.Ban, u types.UserAPI) types.BanAPI {
			return types.BanAPI{
				ID:   b.ID,
				Time: b.Time,
				User: u,
			}
		},
	)
	if err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	return res, nil
}
