/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:11:38
 * @LastEditTime: 2025-05-12 20:57:48
 * @Description: 用户模块业务响应
 */
package service

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/mail"
	"BIT101-GO/pkg/webvpn"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户模块服务
type UserService struct {
	imageSvc   *ImageService
	messageSvc *MessageService
}

// NewUserService 创建用户模块服务
func NewUserService(imageSvc *ImageService, messageSvc *MessageService) *UserService {
	s := &UserService{imageSvc, messageSvc}
	types.RegisterObjHandler(s)
	return s
}

// IsExist 检查用户是否存在
// 实现ObjHandler接口
func (s *UserService) IsExist(id uint) bool {
	_, err := getUser(int(id))
	return err == nil
}

// GetObjType 获取对象类型
// 实现ObjHandler接口
func (s *UserService) GetObjType() string {
	return "user"
}

// GetUid 获取用户ID
// 实现ObjHandler接口
func (s *UserService) GetUid(id uint) (uint, error) {
	userAPI, err := s.GetUserAPI(int(id))
	if err != nil {
		return 0, err
	}
	return uint(userAPI.ID), nil
}

// GetText 获取用户昵称
// 实现ObjHandler接口
func (s *UserService) GetText(id uint) (string, error) {
	userAPI, err := s.GetUserAPI(int(id))
	if err != nil {
		return "", err
	}
	return userAPI.Nickname, nil
}

// LikeHandler 点赞处理
// 实现ObjHandler接口
func (s *UserService) LikeHandler(tx *gorm.DB, id uint, delta int, uid uint) (uint, error) {
	return 0, errors.New("用户不支持点赞Orz")
}

// CommentHandler 评论处理
// 实现ObjHandler接口
func (s *UserService) CommentHandler(tx *gorm.DB, id uint, comment database.Comment, delta int, uid uint) (uint, error) {
	return 0, errors.New("用户不支持评论Orz")
}

// GetAnonymousUserAPI 获取匿名者信息
func (s *UserService) GetAnonymousUserAPI() types.UserAPI {
	return types.UserAPI{
		ID:         -1,
		CreateTime: time.Now(),
		Nickname:   "匿名者",
		Avatar:     s.imageSvc.GetImageAPI(""),
		Motto:      "面对愚昧，匿名者自己也缄口不言。",
		Identity:   database.IdentityNormal.ToIdentity(),
	}
}

// user2UserAPI 数据库用户转API用户
func (s *UserService) user2UserAPI(user database.User) types.UserAPI {
	return types.UserAPI{
		ID:         int(user.ID),
		CreateTime: user.CreatedAt,
		Nickname:   user.Nickname,
		Avatar:     s.imageSvc.GetImageAPI(user.Avatar),
		Motto:      user.Motto,
		Identity:   database.IdentityType(user.Identity).ToIdentity(),
	}
}

// getUser 获取数据库用户信息
func getUser(uid int) (database.User, error) {
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "id = ?", uid).Error; err != nil {
		return database.User{}, errors.New("数据库错误Orz")
	}
	if user.ID == 0 {
		return database.User{}, errors.New("用户不存在Orz")
	}
	return user, nil
}

// getAnonymousName 根据obj和id，hash出独特的匿名者名字
func (s *UserService) GetAnonymousName(uid uint, obj string) string {
	// 使用hash算法生成匿名序号
	hasher := md5.New()
	hasher.Write([]byte(obj + config.Get().Key + fmt.Sprint(uid)))
	hashBytes := hasher.Sum(nil)
	return "匿名者·" + hex.EncodeToString(hashBytes)[:6]
}

// GetUserAPI 获取用户信息
func (s *UserService) GetUserAPI(uid int) (types.UserAPI, error) {
	if uid == -1 {
		return s.GetAnonymousUserAPI(), nil
	}
	user, err := getUser(uid)
	if err != nil {
		return types.UserAPI{}, err
	}
	return s.user2UserAPI(user), nil
}

// GetUserAPIList 批量获取用户信息
func (s *UserService) GetUserAPIList(uidList []int) ([]types.UserAPI, error) {
	uidMap := make(map[int]bool)
	for _, uid := range uidList {
		uidMap[uid] = true
	}
	user_api_map, err := s.GetUserAPIMap(uidMap)
	if err != nil {
		return nil, err
	}
	user_api_list := make([]types.UserAPI, 0, len(uidList))
	for _, uid := range uidList {
		user_api_list = append(user_api_list, user_api_map[uid])
	}
	return user_api_list, nil
}

// GetUserAPIMap 批量获取用户信息
func (s *UserService) GetUserAPIMap(uidMap map[int]bool) (map[int]types.UserAPI, error) {
	out := make(map[int]types.UserAPI)
	uidList := make([]int, 0)
	for uid := range uidMap {
		if uid == -1 {
			out[-1] = s.GetAnonymousUserAPI()
		} else {
			uidList = append(uidList, uid)
		}
	}

	var users []database.User
	if database.DB.Where("id IN ?", uidList).Find(&users).Error != nil {
		return nil, errors.New("数据库错误Orz")
	}
	// 检查是否全部找到
	if len(users) != len(uidList) {
		return nil, errors.New("用户不存在Orz")
	}
	for _, user := range users {
		out[int(user.ID)] = s.user2UserAPI(user)
	}
	return out, nil
}

// FillUsers 批量填充用户信息
// sources 源数据
// uidExtractor 从源数据中提取uid的函数
// targetCreator 创建目标数据的函数
// 接口方法不支持泛型T^T
func FillUsers[S, T any](
	userSvc *UserService,
	sources []S,
	uidExtractor func(S) int,
	targetCreator func(S, types.UserAPI) T,
) ([]T, error) {
	uids := make(map[int]bool)
	for _, source := range sources {
		uid := uidExtractor(source)
		if uid != 0 {
			uids[uid] = true
		}
	}
	userAPIs, err := userSvc.GetUserAPIMap(uids)
	userAPIs[0] = types.UserAPI{}
	if err != nil {
		return nil, err
	}
	targets := make([]T, 0, len(sources))
	for _, source := range sources {
		uid := uidExtractor(source)
		targets = append(targets, targetCreator(source, userAPIs[uid]))
	}
	return targets, nil
}

// Login 登录
func (s *UserService) Login(sid string, pwd string) (fakeCookie string, err error) {
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "sid = ?", sid).Error; err != nil {
		return "", errors.New("数据库错误Orz")
	}
	if user.ID == 0 || user.Password != pwd {
		return "", errors.New("用户名或密码错误Orz")
	}
	fakeCookie = common.NewUserToken(fmt.Sprint(user.ID), config.Get().LoginExpire, config.Get().Key, int(user.Identity))
	return fakeCookie, nil
}

// WebvpnVerifyInit 初始化WebVPN验证 captcha验证码图片base64为空则不需要验证码
func (s *UserService) WebvpnVerifyInit(sid string) (data webvpn.InitLoginReturn, captcha string, err error) {
	data, err = webvpn.InitLogin()
	if err != nil {
		return webvpn.InitLoginReturn{}, "", errors.New("初始化WebVPN验证失败Orz")
	}
	// TODO: captcha
	// needCaptcha, err := webvpn.NeedCaptcha(sid)
	// if err != nil {
	// 	return webvpn.InitLoginReturn{}, "", errors.New("检查是否需要验证失败Orz")
	// }
	// if needCaptcha {
	// 	img, err := webvpn.CaptchaImage(data.Cookie)
	// 	if err != nil {
	// 		return webvpn.InitLoginReturn{}, "", errors.New("获取验证码图片失败Orz")
	// 	}
	// 	captcha = "data:image/png;base64," + base64.StdEncoding.EncodeToString(img)
	// }
	return data, captcha, nil
}

// WebvpnVerify WebVPN验证
func (s *UserService) WebvpnVerify(sid string, salt string, pwd string, execution string, cookie string, captcha string) (token string, code string, err error) {
	err = webvpn.Login(sid, salt, pwd, execution, cookie, captcha)
	if err != nil {
		return "", "", errors.New("统一身份认证失败Orz")
	}
	//生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code = fmt.Sprintf("%06v", rnd.Int31n(1000000))
	token = common.NewUserToken(sid, config.Get().VerifyCodeExpire, config.Get().Key+code, -1)
	return token, code, nil
}

// MailVerify 邮箱验证
func (s *UserService) MailVerify(sid string) (token string, err error) {
	//生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	token = common.NewUserToken(sid, config.Get().VerifyCodeExpire, config.Get().Key+code, -1)
	//发送邮件
	if err := mail.Send(sid+"@bit.edu.cn", "[BIT101]验证码", fmt.Sprintf("【%v】 是你的验证码ヾ(^▽^*)))", code)); err != nil {
		return "", errors.New("发送邮件失败Orz")
	}
	return token, nil
}

// Register 注册
func (s *UserService) Register(token string, code string, pwd string, loginMode bool) (fakeCookie string, err error) {
	// 验证token
	sid, ok, _, _ := common.VeirifyUserToken(token, config.Get().Key+code)
	if !ok {
		return "", errors.New("验证码无效Orz")
	}

	// 查询用户是否已经注册过
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "sid = ?", sid).Error; err != nil {
		return "", errors.New("数据库错误Orz")
	}
	if user.ID == 0 { // 未注册过
		user.Sid = sid
		user.Password = pwd
		user.Motto = "I offer you the BITterness of a man who has looked long and long at the lonely moon." // 默认格言

		// 生成昵称
		user_ := database.User{}
		for {
			nickname := "BIT101-" + uuid.New().String()[:8]
			if err := database.DB.Limit(1).Find(&user_, "nickname = ?", nickname).Error; err != nil {
				return "", errors.New("数据库错误Orz")
			}
			if user_.ID == 0 {
				user.Nickname = nickname
				break
			}
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return "", errors.New("数据库错误Orz")
		}
	} else { // 已经注册过且不处于登录模式则修改密码
		if !loginMode { // 修改密码
			if err := database.DB.Model(&user).Update("password", pwd).Error; err != nil {
				return "", errors.New("数据库错误Orz")
			}
		}
	}
	fakeCookie = common.NewUserToken(fmt.Sprint(user.ID), config.Get().LoginExpire, config.Get().Key, int(user.Identity))
	return fakeCookie, nil
}

// GetInfo 获取用户信息 uid为-1时返回匿名者
func (s *UserService) GetInfo(uid int, myUid uint) (userInfo types.UserInfo, err error) {
	if uid == -1 {
		return types.UserInfo{
			UserAPI:   s.GetAnonymousUserAPI(),
			FollowAPI: types.FollowAPI{}, // 对嵌入字段使用类型名作为字段名
			Own:       false,
		}, nil
	}
	userAPI, err := s.GetUserAPI(uid)
	if err != nil {
		return types.UserInfo{}, err
	}
	followAPI, err := s.GetFollowAPI(uint(uid), myUid)
	if err != nil {
		return types.UserInfo{}, err
	}
	return types.UserInfo{
		UserAPI:   userAPI,
		FollowAPI: followAPI,
		Own:       uint(uid) == myUid,
	}, nil
}

// Deprecated: OldGetInfo 获取用户信息(旧)
func (s *UserService) OldGetInfo(uid int) (types.UserAPI, error) {
	// 匿名用户
	if uid == -1 {
		return s.GetAnonymousUserAPI(), nil
	}
	userAPI, err := s.GetUserAPI(uid)
	if err != nil {
		return types.UserAPI{}, err
	}

	return userAPI, nil
}

// SetInfo 修改用户信息
func (s *UserService) SetInfo(uid int, nickname, avatarMid, motto string) error {
	user, err := getUser(uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("用户不存在Orz")
	}

	if nickname != "" {
		user_ := database.User{}
		if err := database.DB.Limit(1).Find(&user_, "nickname = ?", nickname).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if user_.ID != 0 && user_.ID != user.ID {
			return errors.New("昵称冲突Orz")
		}
		user.Nickname = nickname
	}
	if avatarMid != "" {
		// 验证图片是否存在
		if !s.imageSvc.CheckMid(avatarMid) {
			return errors.New("头像不存在Orz")
		}
		user.Avatar = avatarMid
	}
	if motto != "" {
		user.Motto = motto
	}
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	return nil
}

// GetFollowAPI 获取关注信息
func (s *UserService) GetFollowAPI(followUid, myUid uint) (types.FollowAPI, error) {
	// 统计关注数
	var stats struct {
		FollowingCount int64
		FollowerCount  int64
	}
	if database.DB.Raw(`
        SELECT
            (SELECT COUNT(*) FROM follows WHERE uid = ? AND deleted_at IS NULL) AS following_count,
            (SELECT COUNT(*) FROM follows WHERE follow_uid = ? AND deleted_at IS NULL) AS follower_count
    `, followUid, followUid).Scan(&stats).Error != nil {
		return types.FollowAPI{}, errors.New("数据库错误Orz")
	}

	// 检查关注状态
	type FollowCheck struct {
		Uid       uint
		FollowUid uint
	}
	var followChecks []FollowCheck
	database.DB.Model(&database.Follow{}).
		Select("uid, follow_uid").
		Where("(uid = ? AND follow_uid = ?) OR (uid = ? AND follow_uid = ?)",
			myUid, followUid, followUid, myUid).
		Find(&followChecks)
	following := false
	follower := false
	for _, check := range followChecks {
		if check.Uid == myUid && check.FollowUid == followUid {
			following = true
		}
		if check.Uid == followUid && check.FollowUid == myUid {
			follower = true
		}
	}

	return types.FollowAPI{
		Follower:     follower,
		FollowerNum:  stats.FollowerCount,
		Following:    following,
		FollowingNum: stats.FollowingCount,
	}, nil
}

// Follow 关注
func (s *UserService) Follow(followUid int, myUid uint) error {
	if followUid == -1 {
		return errors.New("不能关注匿名者Orz")
	}
	if followUid == 0 || uint(followUid) == myUid {
		return errors.New("不能关注自己Orz")
	}
	_, err := getUser(followUid)
	if err != nil {
		return errors.New("获取关注用户失败Orz")
	}

	var follow database.Follow
	database.DB.Unscoped().Where("uid = ?", myUid).Where("follow_uid = ?", followUid).Limit(1).Find(&follow)
	if follow.ID == 0 { //新建
		follow = database.Follow{
			Uid:       myUid,
			FollowUid: uint(followUid),
		}
		database.DB.Create(&follow)
		go s.messageSvc.Send(int(follow.Uid), follow.FollowUid, "", "follow", "", "")
	} else if follow.DeletedAt.Valid { //取消删除
		follow.DeletedAt = gorm.DeletedAt{}
		database.DB.Unscoped().Save(&follow)
		go s.messageSvc.Send(int(follow.Uid), follow.FollowUid, "", "follow", "", "")
	} else { //删除
		database.DB.Delete(&follow)
	}
	return nil
}

// GetFollowList 获取关注列表
func (s *UserService) GetFollowList(uid uint, page uint) ([]types.UserAPI, error) {
	var followList []database.Follow
	if database.DB.Where("uid = ?", uid).Order("updated_at DESC").Offset(int(page*config.Get().FollowPageSize)).Limit(int(config.Get().FollowPageSize)).Find(&followList).Error != nil {
		return nil, errors.New("数据库错误Orz")
	}
	usersIds := make([]int, 0, len(followList))
	for _, follow := range followList {
		usersIds = append(usersIds, int(follow.FollowUid))
	}
	users, err := s.GetUserAPIList(usersIds)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetFansList 获取粉丝列表
func (s *UserService) GetFansList(uid uint, page uint) ([]types.UserAPI, error) {
	var followList []database.Follow
	if database.DB.Where("follow_uid = ?", uid).Order("updated_at DESC").Offset(int(page*config.Get().FollowPageSize)).Limit(int(config.Get().FollowPageSize)).Find(&followList).Error != nil {
		return nil, errors.New("数据库错误Orz")
	}
	usersIds := make([]int, 0, len(followList))
	for _, follow := range followList {
		usersIds = append(usersIds, int(follow.Uid))
	}
	users, err := s.GetUserAPIList(usersIds)
	if err != nil {
		return nil, err
	}
	return users, nil
}
