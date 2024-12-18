package search

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

var client *meilisearch.Client

// createAndConfigureIndex 创建并配置索引
func createAndConfigureIndex(uid, primaryKey string, sortAttr []string, filterAttr []string, searchAttr []string) {
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
	index, _ = indexToUpdate.UpdateSortableAttributes(&sortAttr)
	client.WaitForTask(index.TaskUID)
	index, _ = indexToUpdate.UpdateFilterableAttributes(&filterAttr)
	client.WaitForTask(index.TaskUID)
	index, _ = indexToUpdate.UpdateSearchableAttributes(&searchAttr)
	client.WaitForTask(index.TaskUID)
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
func Search(result interface{}, indexName, query string, page, limit uint, sort, filter []string) error {
	// 构造搜索请求
	searchRequest := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(page * limit),
		Sort:   sort,
		Filter: filter,
	}

	// 执行搜索
	response, err := client.Index(indexName).Search(query, searchRequest)
	if err != nil {
		return fmt.Errorf("[Search] 搜索失败: %w", err)
	}

	// 检查结果是否为空
	if len(response.Hits) == 0 {
		return nil
	}

	// 处理搜索结果
	hitsJSON, err := json.Marshal(response.Hits)
	if err != nil {
		return fmt.Errorf("[Search] 编码搜索结果为JSON失败: %w", err)
	}

	if err := json.Unmarshal(hitsJSON, result); err != nil {
		return fmt.Errorf("[Search] 解码搜索结果为目标结构失败: %w", err)
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

// Sync 同步
func Sync(time_after time.Time) {
	dataSlice_course := []database.Course{}
	data_update_course := []database.Course{}
	data_delete_course := []string{}
	q := database.DB.Model(dataSlice_course).Unscoped().Where("updated_at > ?", time_after).Or("deleted_at > ?", time_after)
	if err := q.Find(&dataSlice_course).Error; err != nil {
		fmt.Printf("[Search] %s同步失败\n", "course")
	}
	for _, v := range dataSlice_course {
		if v.DeletedAt.Valid {
			data_delete_course = append(data_delete_course, strconv.Itoa(int(v.ID)))
		} else {
			data_update_course = append(data_update_course, v)
		}
	}
	if err := addDocumentToMeiliSearch("course", data_update_course); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "course")
	}
	if err := deleteDocumentFromMeiliSearch("course", data_delete_course); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "course")
	}

	dataSlice_paper := []database.Paper{}
	data_update_paper := []database.Paper{}
	data_delete_paper := []string{}
	q = database.DB.Model(dataSlice_paper).Unscoped().Where("updated_at > ?", time_after).Or("deleted_at > ?", time_after)
	if err := q.Find(&dataSlice_paper).Error; err != nil {
		fmt.Printf("[Search] %s同步失败\n", "paper")
	}
	for _, v := range dataSlice_paper {
		if v.DeletedAt.Valid {
			data_delete_paper = append(data_delete_paper, strconv.Itoa(int(v.ID)))
		} else {
			data_update_paper = append(data_update_paper, v)
		}
	}
	if err := addDocumentToMeiliSearch("paper", data_update_paper); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "paper")
	}
	if err := deleteDocumentFromMeiliSearch("paper", data_delete_paper); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "paper")
	}

	dataSlice_poster := []database.Poster{}
	data_update_poster := []database.Poster{}
	data_delete_poster := []string{}
	q = database.DB.Model(dataSlice_poster).Unscoped().Where("updated_at > ?", time_after).Or("deleted_at > ?", time_after)
	if err := q.Find(&dataSlice_poster).Error; err != nil {
		fmt.Printf("[Search] %s同步失败\n", "poster")
	}
	for _, v := range dataSlice_poster {
		if v.DeletedAt.Valid {
			data_delete_poster = append(data_delete_poster, strconv.Itoa(int(v.ID)))
		} else {
			data_update_poster = append(data_update_poster, v)
		}
	}
	if err := addDocumentToMeiliSearch("poster", data_update_poster); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "poster")
	}
	if err := deleteDocumentFromMeiliSearch("poster", data_delete_poster); err != nil {
		fmt.Printf("[Search] %s同步失败\n", "poster")
	}
}

// Init 初始化
func Init() {
	client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.Config.Meilisearch.Url,
		APIKey: config.Config.Meilisearch.MasterKey,
	})
	// 可排序参数
	sortAttrCourse := []string{"comment_num", "like_num", "rate", "update_time"}
	sortAttrPaper := []string{"like_num", "update_time"}
	sortAttrPoster := []string{"like_num", "comment_num", "create_time"}
	// 可筛选参数
	filterAttrCourse := []string{}
	filterAttrPaper := []string{}
	filterAttrPoster := []string{"uid", "public", "anonymous"}
	// 可搜索参数
	searchAttrCourse := []string{"name", "number", "teachers_name", "teachers_number", "teachers"}
	searchAttrPaper := []string{"id", "title", "intro", "content"}
	searchAttrPoster := []string{"title", "text", "tags"}

	createAndConfigureIndex("course", "id", sortAttrCourse, filterAttrCourse, searchAttrCourse)
	createAndConfigureIndex("paper", "id", sortAttrPaper, filterAttrPaper, searchAttrPaper)
	createAndConfigureIndex("poster", "id", sortAttrPoster, filterAttrPoster, searchAttrPoster)

	// 与pg数据库同步
	importData("course", &[]database.Course{})
	importData("paper", &[]database.Paper{})
	importData("poster", &[]database.Poster{})
}
