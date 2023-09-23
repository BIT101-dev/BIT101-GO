/*
 * @Author: flwfdd
 * @Date: 2023-09-23 20:14:40
 * @LastEditTime: 2023-09-23 23:41:58
 * @Description: 获取历史均分_(:з」∠)_
 */
package other

import (
	"BIT101-GO/controller/webvpn"
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// 获取课程历史并写入数据库
func GetCourseHistory(start_year_s string, end_year_s string, webvpn_cookie string) {
	// 解析参数
	start_year, err1 := strconv.Atoi(start_year_s)
	end_year, err2 := strconv.Atoi(end_year_s)
	if err1 != nil || err2 != nil {
		panic("invalid year")
	}

	// 获取所有课程编号
	config.Init()
	database.Init()
	course_numbers := []string{}
	database.DB.Model(&database.Course{}).Distinct("number").Find(&course_numbers)
	fmt.Println("共", len(course_numbers), "门课程")

	// 初始化并发
	ch := make(chan *database.CourseHistory, 24)
	ch_count := 0

	// 获取成绩
	for year := start_year; year < end_year; year++ {
		for i := 1; i <= 2; i++ {
			term := fmt.Sprintf("%d-%d-%d", year, year+1, i)
			fmt.Println("正在获取", term, "学期的课程历史")
			// 初始化当前已有的课程历史
			var db_histories []database.CourseHistory
			database.DB.Find(&db_histories)
			hisotry_exist := make(map[string]bool)
			for _, db_history := range db_histories {
				hisotry_exist[db_history.Number+" "+db_history.Term] = true
			}

			// 爬虫，启动！
			for _, course_number := range course_numbers {
				// 检查是否已经存在
				if hisotry_exist[course_number+" "+term] {
					continue
				}
				go func(course_number string, term string) {
					score, _ := getCourseHistoryItem(course_number, term, webvpn_cookie)
					ch <- score
				}(course_number, term)
				ch_count++
			}

			// 等待并发结束
			for ch_count > 0 {
				score := <-ch
				if score != nil {
					database.DB.Where("number = ? AND term = ?", score.Number, score.Term).FirstOrCreate(score)
					fmt.Println(ch_count, score.Term, score.Number)
				}
				ch_count--
			}
		}
	}
}

// 获取单个学期单个课程信息
func getCourseHistoryItem(course_number string, term string, webvpn_cookie string) (*database.CourseHistory, error) {
	mp, err := webvpn.GetCourseHistory(course_number, term, webvpn_cookie)
	if err != nil {
		return nil, err
	}
	avg_score, err1 := strconv.ParseFloat(mp["平均分"], 64)
	max_score, err2 := strconv.ParseFloat(mp["最高分"], 64)
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindStringSubmatch(mp["学习人数"])
	people_num, err3 := strconv.ParseUint(match[1], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || people_num == 0 {
		return nil, errors.New("invalid score")
	}
	return &database.CourseHistory{
		Number:    course_number,
		Term:      term,
		AvgScore:  avg_score,
		MaxScore:  max_score,
		PeopleNum: uint(people_num),
	}, nil
}
