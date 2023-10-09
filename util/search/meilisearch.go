package search

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"encoding/json"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
)

var client *meilisearch.Client

// ImportCourse 将Course表中的数据导入到MeiliSearch中
func ImportCourse() {
	var courses []database.Course
	q := database.DB.Model(&database.Course{})
	if err := q.Find(&courses).Error; err != nil {
		panic(err)
	}

	var coursesData []map[string]interface{}
	for _, course := range courses {
		courseMap := database.StructToMap(course)
		coursesData = append(coursesData, courseMap)
	}

	err := addDocumentToMeiliSearch(client, "course", coursesData)
	if err != nil {
		fmt.Println("Course导入失败")
		return
	}
	fmt.Println("Course导入完成")
}

// ImportPaper 将Paper表中的数据导入到MeiliSearch中
func ImportPaper() {
	var papers []database.Paper
	q := database.DB.Model(&database.Paper{})
	if err := q.Find(&papers).Error; err != nil {
		panic(err)
	}

	var papersData []map[string]interface{}
	for _, paper := range papers {
		courseMap := database.StructToMap(paper)
		papersData = append(papersData, courseMap)
	}

	err := addDocumentToMeiliSearch(client, "paper", papersData)
	if err != nil {
		fmt.Println("Paper导入失败")
		return
	}
	fmt.Println("Paper导入完成")
}

// addDocumentToMeiliSearch 将数据添加到MeiliSearch中
func addDocumentToMeiliSearch(client *meilisearch.Client, indexName string, documents []map[string]interface{}) error {
	if len(documents) == 0 {
		return nil
	}
	index, err := client.GetIndex(indexName)
	if err != nil {
		fmt.Println("No such index:", err)
		return err
	}
	info, _ := index.AddDocuments(documents)
	task, _ := client.WaitForTask(info.TaskUID)
	fmt.Println("addDocument task_uid:", info.TaskUID, ",task_status:", task.Status)
	if task.Status == "failed" {
		return fmt.Errorf("task failed")
	}
	return nil
}

// 将数据从MeiliSearch中删除
func deleteDocumentFromMeiliSearch(client *meilisearch.Client, indexName string, ids []string) error {
	index, err := client.GetIndex(indexName)
	if err != nil {
		fmt.Println("No such index:", err)
		return err
	}
	info, _ := index.DeleteDocuments(ids)
	task, _ := client.WaitForTask(info.TaskUID)
	fmt.Println("deleteDocument task_uid:", info.TaskUID, ",task_status:", task.Status)
	if task.Status == "failed" {
		return fmt.Errorf("task failed")
	}
	return nil
}

// Search 搜索
func Search(result interface{}, indexName string, query string, sort []string, page int64) error {
	response, err := client.Index(indexName).Search(query, &meilisearch.SearchRequest{
		Limit:  int64(config.Config.PaperPageSize),
		Offset: page * int64(config.Config.PaperPageSize),
		Sort:   sort,
	})
	if err != nil {
		return err
	}
	hits := response.Hits
	if len(hits) == 0 {
		return nil
	}

	resultJSON, _ := json.Marshal(hits)
	if err := json.Unmarshal(resultJSON, result); err != nil {
		fmt.Println("解码JSON失败:", err)
		return err
	}
	return nil
}

// Update 更新记录（add、update）
func Update(indexName string, documents []map[string]interface{}) error {
	return addDocumentToMeiliSearch(client, indexName, documents)
}

// Delete 删除记录
func Delete(indexName string, ids []string) error {
	return deleteDocumentFromMeiliSearch(client, indexName, ids)
}

// Init 初始化
func Init() {
	client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   "http://localhost:7700",
		APIKey: config.Config.SearchApiKey,
	})
	// 创建index
	info1, _ := client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "course",
		PrimaryKey: "id",
	})
	info2, _ := client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "paper",
		PrimaryKey: "id",
	})
	client.WaitForTask(info1.TaskUID)
	client.WaitForTask(info2.TaskUID)

	// 设置sort
	courseIndex, _ := client.GetIndex("course")
	paperIndex, _ := client.GetIndex("paper")
	sortableAttributes := []string{"comment_num", "like_num", "rate", "update_time"}
	courseIndex.UpdateSortableAttributes(&sortableAttributes)
	sortableAttributes = []string{"like_num", "update_time"}
	paperIndex.UpdateSortableAttributes(&sortableAttributes)

	// 与pg数据库同步
	ImportCourse()
	ImportPaper()
}

// Test 测试
func Test() {

}
