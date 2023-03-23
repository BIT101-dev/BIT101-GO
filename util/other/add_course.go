/*
 * @Author: flwfdd
 * @Date: 2023-03-23 14:59:32
 * @LastEditTime: 2023-03-23 22:11:56
 * @Description: _(:з」∠)_
 */
package other

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/nlp"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

func AddCourse() {
	config.Init()
	database.Init()

	csv_paths := []string{
		"./data/course/2020-2021-1.csv",
		"./data/course/2020-2021-2.csv",
		"./data/course/2021-2022-1.csv",
		"./data/course/2021-2022-2.csv",
		"./data/course/2022-2023-1.csv",
	}
	for _, csv_path := range csv_paths {
		addFromCsv(csv_path)
	}
}

func addFromCsv(csv_path string) {
	csv_byte, err := ioutil.ReadFile(csv_path)
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

		name_list := strings.Split(course["上课教师"], ",")
		number_list := strings.Split(course["教师号"], ",")
		var db_courses []database.Course
		database.DB.Where("number = ?", course["课程号"]).Find(&db_courses)
		var new_flag = true
		if len(db_courses) > 0 {
			for _, db_course := range db_courses {
				if db_course.Name != course["课程名"] {
					fmt.Printf("课程名不一致: %s %s\n", db_course.Name, course["课程名"])
				}
				number_list_ := strings.Split(db_course.TeachersNumber, ",")
				sort.Strings(number_list)
				sort.Strings(number_list_)
				if strings.Join(number_list, ",") == strings.Join(number_list_, ",") {
					new_flag = false
					break
				}
			}
		}
		if new_flag {
			teaches := getTeachers(name_list, number_list)
			tsv := database.Tsvector{}
			for _, teacher := range teaches {
				tsv.D = append(tsv.D, nlp.CutAll(teacher.Name)...)
				tsv.A = append(tsv.A, teacher.Number)
			}
			tsv.D = append(tsv.D, nlp.CutAll(course["课程名"])...)
			tsv.A = append(tsv.A, course["课程号"])
			db_course := database.Course{
				Name:           course["课程名"],
				Number:         course["课程号"],
				TeachersName:   course["上课教师"],
				TeachersNumber: course["教师号"],
				Teachers:       teaches,
				Tsv:            tsv,
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

func getTeachers(name_list []string, number_list []string) []database.Teacher {
	var db_teachers []database.Teacher
	for ind, _ := range number_list {
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
