/*
 * @Author: flwfdd
 * @Date: 2023-03-23 14:59:32
 * @LastEditTime: 2023-10-10 19:55:03
 * @Description: _(:з」∠)_
 */
package other

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

	fmt.Printf("Detected %d file(s) in %s:\n", len(csv_paths), path)
	for _, csv_path := range csv_paths {
		fmt.Println(csv_path)
	}
	fmt.Println()

	// 导入课程数据
	config.Init()
	database.Init()
	for _, csv_path := range csv_paths {
		importFromCsv(csv_path)
	}
}

func importFromCsv(csv_path string) {
	fmt.Println("开始导入课程数据:", csv_path)
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
		course := make(map[string]string)
		for ind, s := range course_list {
			course[courses[0][ind]] = s
		}

		course["课程名"] = strings.ReplaceAll(course["课程名"], "/", "_")
		name_list := strings.Split(course["上课教师"], ",")
		number_list := strings.Split(course["教师号"], ",")
		var db_courses []database.Course
		database.DB.Where("number = ?", course["课程号"]).Find(&db_courses)
		var new_flag = true // 是否为新课程
		for _, db_course := range db_courses {
			if db_course.Name != course["课程名"] {
				fmt.Printf("课程名不一致: %s %s\n", db_course.Name, course["课程名"])
			}
			teacher_number_list := strings.Split(db_course.TeachersNumber, ",")
			sort.Strings(number_list)
			sort.Strings(teacher_number_list)
			// 教师号相同则不是新课程
			if strings.Join(number_list, ",") == strings.Join(teacher_number_list, ",") {
				new_flag = false
				break
			}
		}

		// 加入新课程
		if new_flag {
			teaches := getTeachers(name_list, number_list)
			db_course := database.Course{
				Name:           course["课程名"],
				Number:         course["课程号"],
				TeachersName:   course["上课教师"],
				TeachersNumber: course["教师号"],
				Teachers:       teaches,
			}
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
}

// 打包教师信息
func getTeachers(name_list []string, number_list []string) []database.Teacher {
	var db_teachers []database.Teacher
	for ind := range number_list {
		var db_teacher database.Teacher
		database.DB.Where("number = ?", number_list[ind]).Limit(1).Find(&db_teacher)
		if db_teacher.ID == 0 {
			db_teacher = database.Teacher{
				Name:   name_list[ind],
				Number: number_list[ind],
			}
			database.DB.Create(&db_teacher)
		}
		db_teachers = append(db_teachers, db_teacher)
	}
	return db_teachers
}
