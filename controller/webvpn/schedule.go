/*
 * @Author: flwfdd
 * @Date: 2023-03-29 16:24:34
 * @LastEditTime: 2023-03-29 20:20:32
 * @Description: _(:з」∠)_
 */
package webvpn

import (
	"BIT101-GO/util/request"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ScheduleItem struct {
	KCM              string `json:"KCM"`              //课程名
	SKJS             string `json:"SKJS"`             //授课教师 逗号分隔
	JASMC            string `json:"JASMC"`            //教室
	YPSJDD           string `json:"YPSJDD"`           //上课时空描述
	SKZC             string `json:"SKZC"`             //上课周次 是一个01串 1表示上课 0表示不上课
	SKXQ             int    `json:"SKXQ"`             //星期几
	KSJC             int    `json:"KSJC"`             //开始节次
	JSJC             int    `json:"JSJC"`             //结束节次
	XXXQMC           string `json:"XXXQMC"`           //校区名
	KCH              string `json:"KCH"`              //课程号
	XF               int    `json:"XF"`               //学分
	KCXZDM_DISPLAY   string `json:"KCXZDM_DISPLAY"`   //课程性质 必修选修什么的
	XGXKLBDM_DISPLAY string `json:"XGXKLBDM_DISPLAY"` //课程类别 文化课实践课什么的
	DWDM_DISPLAY     string `json:"DWDM_DISPLAY"`     //开课单位
}

// 获取课表返回结构结构
type GetScheduleResponse struct {
	Term     string         `json:"term"`     //学期
	FirstDay string         `json:"firstDay"` //第一天
	Data     []ScheduleItem `json:"data"`     //课表数据
}

func GetSchedule(cookie string, term string) (GetScheduleResponse, error) {
	base_url := "https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421faef5b842238695c720999bcd6572a216b231105adc27d"
	res, err := request.Get(base_url+"/jwapp/sys/funauthapp/api/getAppConfig/wdkbby-5959167891382285.do", map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return GetScheduleResponse{}, errors.New("get schedule init error")
	}
	if strings.Contains(res.Text, "帐号登录或动态码登录") {
		return GetScheduleResponse{}, ErrCookieInvalid
	}

	// 设置语言
	res, err = request.Get(base_url+"/jwapp/i18n.do?appName=wdkbby&EMAP_LANG=zh", map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return GetScheduleResponse{}, errors.New("get schedule lang error")
	}

	// 获取当前学期
	if term == "" {
		res, err = request.Get(base_url+"/jwapp/sys/wdkbby/modules/jshkcb/dqxnxq.do", map[string]string{"Cookie": cookie})
		if err != nil || res.Code != 200 {
			return GetScheduleResponse{}, errors.New("get schedule term error")
		}

		var now_term_json struct {
			Datas struct {
				DQXNXQ struct {
					Rows []struct {
						DM string `json:"DM"`
					} `json:"rows"`
				} `json:"dqxnxq"`
			} `json:"datas"`
		}
		err = json.Unmarshal([]byte(res.Text), &now_term_json)
		if err != nil || len(now_term_json.Datas.DQXNXQ.Rows) == 0 {
			return GetScheduleResponse{}, err
		}
		term = now_term_json.Datas.DQXNXQ.Rows[0].DM
	}

	// 获取学期第一天日期
	first_day := ""
	res, err = request.PostForm(base_url+"/jwapp/sys/wdkbby/wdkbByController/cxzkbrq.do", map[string]string{"requestParamStr": fmt.Sprintf(`{"XNXQDM":"%v","ZC":"1"}`, term)}, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return GetScheduleResponse{}, errors.New("get schedule first day error")
	}
	var first_day_json struct {
		Data []struct {
			XQ int    `json:"XQ"`
			RQ string `json:"RQ"`
		} `json:"data"`
	}
	err = json.Unmarshal([]byte(res.Text), &first_day_json)
	if err != nil {
		return GetScheduleResponse{}, err
	}
	for _, v := range first_day_json.Data {
		if v.XQ == 1 {
			first_day = v.RQ
			break
		}
	}

	// 获取课表
	res, err = request.PostForm(base_url+"/jwapp/sys/wdkbby/modules/xskcb/cxxszhxqkb.do", map[string]string{"XNXQDM": term}, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return GetScheduleResponse{}, errors.New("get schedule error")
	}
	var schedule_json struct {
		Datas struct {
			CXXSZHXQKB struct {
				Rows []ScheduleItem `json:"rows"`
			} `json:"cxxszhxqkb"`
		} `json:"datas"`
	}
	err = json.Unmarshal([]byte(res.Text), &schedule_json)
	if err != nil {
		return GetScheduleResponse{}, err
	}

	return GetScheduleResponse{
		Term:     term,
		FirstDay: first_day,
		Data:     schedule_json.Datas.CXXSZHXQKB.Rows,
	}, nil
}
