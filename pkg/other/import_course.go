/*
 * @Author: flwfdd
 * @Date: 2023-03-23 14:59:32
 * @LastEditTime: 2025-03-11 18:45:59
 * @Description: 导入课程数据_(:з」∠)_
 */
package other

import (
	"BIT101-GO/database"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 原始课程数据 教师名和教师号未对应
type OriCourse struct {
	name                string
	number              string
	teacher_name_list   []string
	teacher_number_list []string
}

func ImportCourse(path string) {
	// 获取 path/*.csv
	csv_paths := []string{}
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Failed to get course file in ", path)
		return
	}
	// 筛选出.csv文件
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			csv_paths = append(csv_paths, filepath.Join(path, file.Name()))
		}
	}

	fmt.Printf("Detected %d csv file(s) in %s:\n", len(csv_paths), path)
	for _, csv_path := range csv_paths {
		fmt.Println(csv_path)
	}
	fmt.Println()

	// 导入课程数据
	database.Init()
	ori_courses := []OriCourse{}
	for _, csv_path := range csv_paths {
		ori_courses = append(ori_courses, importFromCsv(csv_path)...)
	}
	fmt.Printf("Detected %d origin course(s) in total:\n", len(ori_courses))

	// 推断教师信息
	number2name := inferTeachers(ori_courses)

	// 导入课程数据
	importToDatabase(ori_courses, number2name)
}

// 导入课程数据
func importToDatabase(ori_courses []OriCourse, number2name map[string]string) {
	for _, ori_course := range ori_courses {
		var db_courses []database.Course
		database.DB.Where("number = ?", ori_course.number).Find(&db_courses)
		var new_flag = true // 是否为新课程
		for _, db_course := range db_courses {
			if db_course.Name != ori_course.name {
				fmt.Printf("Different course name: %s %s\n", db_course.Name, ori_course.name)
				db_course.Name = ori_course.name
				database.DB.UpdateColumns(&db_course)
			}
			teacher_number_list := strings.Split(db_course.TeachersNumber, ",")
			sort.Strings(ori_course.teacher_number_list)
			sort.Strings(teacher_number_list)
			// 教师号相同则不是新课程
			if strings.Join(ori_course.teacher_number_list, ",") == strings.Join(teacher_number_list, ",") {
				new_flag = false
				break
			}
		}

		// 加入新课程
		if new_flag {
			db_course := database.Course{
				Name:           ori_course.name,
				Number:         ori_course.number,
				TeachersName:   "",
				TeachersNumber: strings.Join(ori_course.teacher_number_list, ","),
				Teachers:       []database.Teacher{},
			}
			db_course.UpdatedAt = time.Date(2022, 8, 1, 5, 0, 0, 0, time.UTC)
			database.DB.Create(&db_course)

			var readme database.CourseUploadReadme
			database.DB.Where("course_number = ?", db_course.Number).Limit(1).Find(&readme)
			if readme.ID == 0 {
				readme = database.CourseUploadReadme{
					CourseNumber: db_course.Number,
				}
				database.DB.Create(&readme)
			}
		}
	}

	// 完善课程教师信息
	var db_courses []database.Course
	database.DB.Find(&db_courses)
	for _, db_course := range db_courses {
		teacher_number_list := strings.Split(db_course.TeachersNumber, ",")
		teacher_name_list := []string{}
		for _, teacher_number := range teacher_number_list {
			teacher_name_list = append(teacher_name_list, number2name[teacher_number])
		}
		db_course.TeachersName = strings.Join(teacher_name_list, ",")
		teachers := getTeachers(teacher_number_list, number2name)
		db_course.Teachers = teachers
		database.DB.UpdateColumns(&db_course)
	}

	// 更新教师信息
	var db_teachers []database.Teacher
	database.DB.Find(&db_teachers)
	for _, db_teacher := range db_teachers {
		if db_teacher.Name != number2name[db_teacher.Number] {
			fmt.Printf("Different teacher name: %s %s %s\n", db_teacher.Number, db_teacher.Name, number2name[db_teacher.Number])
			db_teacher.Name = number2name[db_teacher.Number]
			database.DB.Save(&db_teacher)
		}
	}
}

// 打包教师信息
func getTeachers(number_list []string, number2name map[string]string) []database.Teacher {
	var db_teachers []database.Teacher
	for ind := range number_list {
		var db_teacher database.Teacher
		database.DB.Where("number = ?", number_list[ind]).Limit(1).Find(&db_teacher)
		if db_teacher.ID == 0 {
			db_teacher = database.Teacher{
				Name:   number2name[number_list[ind]],
				Number: number_list[ind],
			}
			database.DB.Create(&db_teacher)
		}
		db_teachers = append(db_teachers, db_teacher)
	}
	return db_teachers
}

// 从csv文件提取课程数据
func importFromCsv(csv_path string) []OriCourse {
	ori_courses := []OriCourse{}
	csv_byte, err := os.ReadFile(csv_path)
	if err != nil {
		panic(err)
	}
	// 读取文件数据
	courses, err := csv.NewReader(strings.NewReader(string(csv_byte))).ReadAll()
	if err != nil {
		panic(err)
	}

	for _, course_list := range courses[1:] {
		ori_course := OriCourse{}
		course := make(map[string]string)
		for ind, s := range course_list {
			course[strings.TrimPrefix(courses[0][ind], "\uFEFF")] = s
		}

		ori_course.name = strings.ReplaceAll(course["课程名"], "/", "_")
		ori_course.number = course["课程号"]
		ori_course.teacher_name_list = strings.Split(course["上课教师"], ",")
		ori_course.teacher_number_list = strings.Split(course["教师号"], ",")
		sort.Strings(ori_course.teacher_number_list)
		ori_courses = append(ori_courses, ori_course)
	}

	return ori_courses
}

// 推断教师信息 匹配教师名和教师号 返回教师号-教师名映射
type inferTeacher struct {
	number     string   // 教师号
	names      []string // 备选教师名
	name_index int      // 当前选择的教师名索引
}

func inferTeachers(ori_courses []OriCourse) map[string]string {
	number2name := make(map[string]string)
	courses1 := []OriCourse{}
	courses2 := []OriCourse{}
	courses1 = append(courses1, ori_courses...)

	// 确定匹配 通过一对一的匹配进行逐步排除
	fmt.Println("Infering teachers round 1")
	uninferred_num := 0
	for i := 0; uninferred_num != len(courses1); i++ {
		uninferred_num = len(courses1)
		fmt.Printf("Remain %d course(s) to infer\n", uninferred_num)
		for _, course := range courses1 {
			number_list, name_list := filterCourse(course, number2name)

			if len(number_list) > 1 {
				// 未完成推断
				courses2 = append(courses2, course)
			} else if len(number_list) == 1 {
				// 推断成功
				if v, ok := number2name[number_list[0]]; ok {
					// 检测冲突
					if v != name_list[0] {
						fmt.Printf("Conflict: %s %s %s\n", number_list[0], v, name_list[0])
						panic("Conflict")
					}
				} else {
					number2name[number_list[0]] = name_list[0]
				}
			}
		}

		courses1 = courses2
		courses2 = []OriCourse{}
	}

	infer_teachers := []inferTeacher{}
	// 交集匹配
	fmt.Println("Infering teachers round 2")
	uninferred_num = 0
	for i := 0; uninferred_num != len(courses1); i++ {
		uninferred_num = len(courses1)
		fmt.Printf("Remain %d course(s) to infer\n", uninferred_num)

		infer_teachers = []inferTeacher{}
		for _, course := range courses1 {
			// 先排除确定匹配
			number_list, name_list := filterCourse(course, number2name)

			// 推断交集
			for _, teacher_number := range number_list {
				// 查找教师号
				teacher_index := 0
				for _, infer_teacher := range infer_teachers {
					if infer_teacher.number == teacher_number {
						break
					}
					teacher_index++
				}
				if teacher_index == len(infer_teachers) {
					// 没有出现过的教师号 所有教师名都是备选教师名
					infer_teachers = append(infer_teachers, inferTeacher{
						number:     teacher_number,
						names:      []string{},
						name_index: 0,
					})
					for _, name := range name_list {
						infer_teachers[teacher_index].names = append(infer_teachers[teacher_index].names, name)
					}
				} else {
					// 出现过的教师号 取交集
					intersection_names := []string{}
					for _, name := range infer_teachers[teacher_index].names {
						for _, name_ := range name_list {
							if name == name_ {
								intersection_names = append(intersection_names, name)
								break
							}
						}
					}
					infer_teachers[teacher_index].names = intersection_names
				}
			}
		}

		// 添加确定匹配
		for _, infer_teacher := range infer_teachers {
			if len(infer_teacher.names) == 1 {
				number2name[infer_teacher.number] = infer_teacher.names[0]
			}
		}

		// 筛选未完成推断的课程
		for _, course := range courses1 {
			for _, teacher_number := range course.teacher_number_list {
				if _, ok := number2name[teacher_number]; !ok {
					// 存在未推断的教师号
					courses2 = append(courses2, course)
					break
				}
			}
		}
		courses1 = courses2
		courses2 = []OriCourse{}
	}

	// 搜索推断 不一定正确但满足当前约束
	fmt.Println("Infering teachers round 3")
	if !searchInferTeachers(courses1, number2name, infer_teachers, 0) {
		panic("Failed to search infer teachers")
	}

	for _, infer_teacher := range infer_teachers {
		number2name[infer_teacher.number] = infer_teacher.names[infer_teacher.name_index]
	}

	fmt.Printf("Infered %d teacher(s) in total\n", len(number2name))
	return number2name
}

// 搜索匹配方案
func searchInferTeachers(courses []OriCourse, number2name map[string]string, infer_teachers []inferTeacher, teacher_index int) bool {
	if teacher_index == len(infer_teachers) {
		return true
	}
	for i := 0; i < len(infer_teachers[teacher_index].names); i++ {
		infer_teachers[teacher_index].name_index = i
		if checkInferTeachers(courses, number2name, infer_teachers[:teacher_index+1]) {
			if searchInferTeachers(courses, number2name, infer_teachers, teacher_index+1) {
				return true
			}
		}
	}
	return false
}

// 检查匹配方案
func checkInferTeachers(courses []OriCourse, number2name map[string]string, infer_teachers []inferTeacher) bool {
	for _, course := range courses {
		// 先排除确定匹配
		number_list, name_list := filterCourse(course, number2name)

		// 排除非确定匹配
		for _, teacher_number := range number_list {
			teacher_index := 0
			for _, infer_teacher := range infer_teachers {
				if infer_teacher.number == teacher_number {
					break
				}
				teacher_index++
			}
			if teacher_index == len(infer_teachers) {
				// 没有出现过的教师号
				continue
			}
			name_index := 0
			for _, name := range name_list {
				if name == infer_teachers[teacher_index].names[infer_teachers[teacher_index].name_index] {
					break
				}
				name_index++
			}
			// 找不到对应的教师名
			if name_index == len(name_list) {
				return false
			}
			// 移除已匹配的教师名
			name_list = removeElement(name_list, name_list[name_index])
		}
	}
	return true
}

// 排除确定匹配
func filterCourse(course OriCourse, number2name map[string]string) ([]string, []string) {
	number_list := []string{}
	name_list := []string{}
	name_list = append(name_list, course.teacher_name_list...)
	for _, teacher_number := range course.teacher_number_list {
		if name, ok := number2name[teacher_number]; ok {
			name_list = removeElement(name_list, name)
		} else {
			number_list = append(number_list, teacher_number)
		}
	}
	if len(number_list) != len(name_list) {
		fmt.Println(course)
		panic("number_list and name_list length not equal")
	}

	return number_list, name_list
}

// 删除元素
func removeElement(l []string, s string) []string {
	for ind, v := range l {
		if v == s {
			return append(l[:ind], l[ind+1:]...)
		}
	}
	return l
}
