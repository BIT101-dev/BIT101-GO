package search

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"encoding/json"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

var client *meilisearch.Client

// createAndConfigureIndex 创建并配置索引
func createAndConfigureIndex(uid, primaryKey string, sortableAttributes []string) {
	index, err := client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        uid,
		PrimaryKey: primaryKey,
	})
	if err != nil {
		fmt.Println("[Search] CreateIndex failed:", err)
		panic(err)
	}
	client.WaitForTask(index.TaskUID)
	indexToUpdate, _ := client.GetIndex(uid)
	indexToUpdate.UpdateSortableAttributes(&sortableAttributes)
}

// importData 导入数据库表到MeiliSearch
func importData(indexName string, dataSlice interface{}) {
	q := database.DB.Model(dataSlice)
	if err := q.Find(dataSlice).Error; err != nil {
		panic(err)
	}

	err := addDocumentToMeiliSearch(indexName, dataSlice)
	if err != nil {
		fmt.Printf("[Search] %s导入失败\n", indexName)
		return
	}
	fmt.Printf("[Search] %s导入完成\n", indexName)
}

// addDocumentToMeiliSearch 将数据添加到MeiliSearch中
func addDocumentToMeiliSearch(indexName string, documents interface{}) error {
	index, err := client.GetIndex(indexName)
	if err != nil {
		fmt.Println("[Search] No such index:", err)
		return err
	}
	info, _ := index.AddDocuments(documents)
	task, _ := client.WaitForTask(info.TaskUID)
	fmt.Println("[Search] Add documents to ", indexName, " task_uid:", info.TaskUID, ",task_status:", task.Status)
	if task.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf("[Search] task %d %s", info.TaskUID, task.Status)
	}
	return nil
}

// deleteDocumentFromMeiliSearch 将数据从MeiliSearch中删除
func deleteDocumentFromMeiliSearch(indexName string, ids []string) error {
	index, err := client.GetIndex(indexName)
	if err != nil {
		fmt.Println("[Search] No such index:", err)
		return err
	}
	info, _ := index.DeleteDocuments(ids)
	task, _ := client.WaitForTask(info.TaskUID)
	fmt.Println("[Search] deleteDocument task_uid:", info.TaskUID, ",task_status:", task.Status)
	if task.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf("[Search] task %d %s", info.TaskUID, task.Status)
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
		fmt.Println("[Search] 解码JSON失败:", err)
		return err
	}
	return nil
}

// Update 更新记录（add、update）
func Update(indexName string, documents interface{}) error {
	return addDocumentToMeiliSearch(indexName, documents)
}

// Delete 删除记录
func Delete(indexName string, ids []string) error {
	return deleteDocumentFromMeiliSearch(indexName, ids)
}

// Init 初始化
func Init() {
	client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.Config.Meilisearch.Url,
		APIKey: config.Config.Meilisearch.MasterKey,
	})
	sortableAttributesCourse := []string{"comment_num", "like_num", "rate", "update_time"}
	sortableAttributesPaper := []string{"like_num", "update_time"}
	sortableAttributesPost := []string{"like_num", "comment_num", "update_time"}

	createAndConfigureIndex("course", "id", sortableAttributesCourse)
	createAndConfigureIndex("paper", "id", sortableAttributesPaper)
	createAndConfigureIndex("post", "id", sortableAttributesPost)

	// 与pg数据库同步
	importData("course", &[]database.Course{})
	importData("paper", &[]database.Paper{})
	importData("post", &[]database.Post{})
}
