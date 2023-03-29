/*
 * @Author: flwfdd
 * @Date: 2023-03-24 00:32:23
 * @LastEditTime: 2023-03-25 15:19:42
 * @Description: _(:з」∠)_
 */
package other

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/nlp"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var base_path = "./data/json/"

type UserJson struct {
	ID           int    `json:"id"`
	Sid          string `json:"sid"`
	Password     string `json:"password"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Motto        string `json:"motto"`
	Level        int    `json:"level"`
	RegisterTime string `json:"register_time"`
}

func MigrateUser() {
	text, err := os.ReadFile(base_path + "user.json")
	if err != nil {
		panic(err)
	}
	var user_list []UserJson
	err = json.Unmarshal(text, &user_list)
	if err != nil {
		panic(err)
	}

	for _, user := range user_list {
		database.DB.Exec(`INSERT INTO users ("id", "created_at", "updated_at", "sid", "password", "nickname", "avatar", "motto", "level")`+
			`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, user.ID, user.RegisterTime, user.RegisterTime, user.Sid, user.Password, user.Nickname, user.Avatar, user.Motto, user.Level)
	}

	database.DB.Exec(`select setval('users_id_seq',(select max(id) from "users"))`)
}

type PaperJson struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Intro      string `json:"intro"`
	Data       string `json:"data"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	User       int    `json:"user"`
	Anonymous  int    `json:"anonymous"`
	LikeNum    int    `json:"like_num"`
	CommentNum int    `json:"comment_num"`
	Show       int    `json:"show"`
	Owner      int    `json:"owner"`
	Share      int    `json:"share"`
}

func MigratePaper() {
	text, err := os.ReadFile(base_path + "paper.json")
	if err != nil {
		panic(err)
	}
	var paper_list []PaperJson
	err = json.Unmarshal(text, &paper_list)
	if err != nil {
		panic(err)
	}

	for _, paper := range paper_list {
		text, err := nlp.ParseEditorJS(paper.Data)
		if err != nil {
			panic(err)
		}

		t := database.Tsvector{
			B: nlp.CutForSearch(paper.Title),
			C: nlp.CutForSearch(paper.Intro),
			D: nlp.CutForSearch(text),
		}
		tsv := gorm.Expr(
			`setweight(to_tsvector('simple',?),'A') || setweight(to_tsvector('simple',?),'B')  || setweight(to_tsvector('simple',?),'C') || setweight(to_tsvector('simple',?),'D')`,
			strings.Join(t.A, " "), strings.Join(t.B, " "), strings.Join(t.C, " "), strings.Join(t.D, " "),
		)
		if paper.Show == 1 {
			database.DB.Exec(`INSERT INTO papers ("id", "created_at", "updated_at", "deleted_at", "title", "intro", "content", "create_uid", "update_uid", "anonymous", "like_num", "comment_num", "public_edit", "edit_at", "tsv")`+
				`VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, paper.ID, paper.CreateTime, paper.UpdateTime, nil, paper.Title, paper.Intro, paper.Data, paper.Owner, paper.User, paper.Anonymous == 1, paper.LikeNum, paper.CommentNum, paper.Share == 1, paper.UpdateTime, tsv)
		} else {
			database.DB.Exec(`INSERT INTO papers ("id", "created_at", "updated_at", "deleted_at", "title", "intro", "content", "create_uid", "update_uid", "anonymous", "like_num", "comment_num", "public_edit", "edit_at", "tsv")`+
				`VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, paper.ID, paper.CreateTime, paper.UpdateTime, paper.UpdateTime, paper.Title, paper.Intro, paper.Data, paper.Owner, paper.User, paper.Anonymous == 1, paper.LikeNum, paper.CommentNum, paper.Share == 1, paper.UpdateTime, tsv)
		}
	}

	database.DB.Exec(`select setval('papers_id_seq',(select max(id) from "papers"))`)
}

type PaperHistoryJson struct {
	ID         int    `json:"id"`
	PaperId    int    `json:"paper_id"`
	Title      string `json:"title"`
	Intro      string `json:"intro"`
	Data       string `json:"data"`
	CreateTime string `json:"create_time"`
	User       int    `json:"user"`
	Anonymous  int    `json:"anonymous"`
}

func MigratePaperHistory() {
	text, err := os.ReadFile(base_path + "paper_history.json")
	if err != nil {
		panic(err)
	}
	var paper_histories []PaperHistoryJson
	err = json.Unmarshal(text, &paper_histories)
	if err != nil {
		panic(err)
	}

	for _, paper_history := range paper_histories {
		database.DB.Exec(`INSERT INTO paper_histories ("id", "pid", "created_at", "updated_at", "deleted_at", "title", "intro", "content", "uid", "anonymous")`+
			`VALUES (?,?,?,?,?,?,?,?,?,?)`, paper_history.ID, paper_history.PaperId, paper_history.CreateTime, paper_history.CreateTime, nil, paper_history.Title, paper_history.Intro, paper_history.Data, paper_history.User, paper_history.Anonymous == 1)
	}

	database.DB.Exec(`select setval('paper_histories_id_seq',(select max(id) from "paper_histories"))`)
}

type LikeJson struct {
	ID   int    `json:"id"`
	User int    `json:"user"`
	Obj  string `json:"obj"`
	Show int    `json:"show"`
	Time string `json:"time"`
}

func MigrateLike() {
	text, err := os.ReadFile(base_path + "like.json")
	if err != nil {
		panic(err)
	}
	var like_list []LikeJson
	err = json.Unmarshal(text, &like_list)
	if err != nil {
		panic(err)
	}

	for _, like := range like_list {
		if like.Show == 0 {
			database.DB.Exec(`INSERT INTO likes ("id", "created_at", "updated_at", "deleted_at", "uid", "obj")`+
				`VALUES (?,?,?,?,?,?)`, like.ID, like.Time, like.Time, like.Time, like.User, like.Obj)
		} else {
			database.DB.Exec(`INSERT INTO likes ("id", "created_at", "updated_at", "deleted_at", "uid", "obj")`+
				`VALUES (?,?,?,?,?,?)`, like.ID, like.Time, like.Time, nil, like.User, like.Obj)
		}
	}

	database.DB.Exec(`select setval('likes_id_seq',(select max(id) from "likes"))`)
}

type CommentJson struct {
	ID         int    `json:"id"`
	User       int    `json:"user"`
	Obj        string `json:"obj"`
	Text       string `json:"text"`
	Show       int    `json:"show"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	Anonymous  int    `json:"anonymous"`
	LikeNum    int    `json:"like_num"`
	CommentNum int    `json:"comment_num"`
	ReplyUser  int    `json:"reply_user"`
	Rate       int    `json:"rate"`
}

func MigrateComment() {
	text, err := os.ReadFile(base_path + "comment.json")
	if err != nil {
		panic(err)
	}
	var comment_list []CommentJson
	err = json.Unmarshal(text, &comment_list)
	if err != nil {
		panic(err)
	}

	for _, comment := range comment_list {
		if comment.Show == 0 {
			database.DB.Exec(`INSERT INTO comments ("id", "created_at", "updated_at", "deleted_at", "uid", "obj", "text", "anonymous", "like_num", "comment_num", "reply_uid", "rate")`+
				`VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, comment.ID, comment.CreateTime, comment.UpdateTime, comment.UpdateTime, comment.User, comment.Obj, comment.Text, comment.Anonymous == 1, comment.LikeNum, comment.CommentNum, comment.ReplyUser, comment.Rate)
		} else {
			database.DB.Exec(`INSERT INTO comments ("id", "created_at", "updated_at", "deleted_at", "uid", "obj", "text", "anonymous", "like_num", "comment_num", "reply_uid", "rate")`+
				`VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, comment.ID, comment.CreateTime, comment.UpdateTime, nil, comment.User, comment.Obj, comment.Text, comment.Anonymous == 1, comment.LikeNum, comment.CommentNum, comment.ReplyUser, comment.Rate)
		}
	}

	database.DB.Exec(`select setval('comments_id_seq',(select max(id) from "comments"))`)
}

type ImageJson struct {
	ID   string `json:"id"`
	Size int    `json:"size"`
	Name string `json:"name"`
	Time string `json:"time"`
	User int    `json:"user"`
}

type ImageArray []ImageJson

func (array ImageArray) Len() int {
	return len(array)
}

func (array ImageArray) Less(i, j int) bool {
	return array[i].Time < array[j].Time
}

func (array ImageArray) Swap(i, j int) {
	array[i], array[j] = array[j], array[i]
}

func MigrateImage() {
	text, err := os.ReadFile(base_path + "image.json")
	if err != nil {
		panic(err)
	}
	var image_list []ImageJson
	err = json.Unmarshal(text, &image_list)
	if err != nil {
		panic(err)
	}

	sort.Sort(ImageArray(image_list))

	for _, image := range image_list {
		database.DB.Exec(`INSERT INTO images ("mid", "created_at", "updated_at", "deleted_at", "uid", "size")`+
			`VALUES (?,?,?,?,?,?)`, image.ID, image.Time, image.Time, nil, image.User, image.Size)
	}

	database.DB.Exec(`select setval('images_id_seq',(select max(id) from "images"))`)
}

type CourseJson struct {
	ID             int     `json:"id"`
	Number         string  `json:"number"`
	Name           string  `json:"name"`
	LikeNum        int     `json:"like_num"`
	CommentNum     int     `json:"comment_num"`
	RateSum        int     `json:"rate_sum"`
	Rate           float64 `json:"rate"`
	TeachersName   string  `json:"teachers_name"`
	TeachersNumber string  `json:"teachers_number"`
	UpdateTime     string  `json:"update_time"`
}

func MigrateCourse() {
	AddCourse()

	text, err := os.ReadFile(base_path + "course.json")
	if err != nil {
		panic(err)
	}
	var course_list []CourseJson
	err = json.Unmarshal(text, &course_list)
	if err != nil {
		panic(err)
	}

	for _, course := range course_list {
		database.DB.Exec(`UPDATE courses SET "like_num"=?, "comment_num"=?, "rate_sum"=?, "rate"=?, "created_at"=?, "updated_at"=? WHERE "id"=?`,
			course.LikeNum, course.CommentNum, course.RateSum, course.Rate, course.UpdateTime, course.UpdateTime, course.ID)
	}
}

type VariableJson struct {
	ID   int    `json:"id"`
	Obj  string `json:"obj"`
	Data string `json:"data"`
}

func MigrateVariable() {
	text, err := os.ReadFile(base_path + "variable.json")
	if err != nil {
		panic(err)
	}
	var variable_json []VariableJson
	err = json.Unmarshal(text, &variable_json)
	if err != nil {
		panic(err)
	}

	for _, variable := range variable_json {
		database.DB.Exec(`INSERT INTO variables ("id", "created_at", "updated_at", "deleted_at", "obj", "data")`+
			`VALUES (?,?,?,?,?,?)`, variable.ID, time.Now(), time.Now(), nil, variable.Obj, variable.Data)
	}

	database.DB.Exec(`select setval('variables_id_seq',(select max(id) from "variables"))`)
}

func Migrate() {
	config.Init()
	database.Init()
	MigrateUser()
	MigratePaper()
	MigrateLike()
	MigratePaperHistory()
	MigrateComment()
	MigrateImage()
	MigrateCourse()
	MigrateVariable()
}
