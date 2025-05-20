/*
 * @Author: flwfdd
 * @Date: 2025-03-04 22:51:41
 * @LastEditTime: 2025-03-18 18:16:53
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/config"
	"BIT101-GO/database"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchService struct {
	client        *meilisearch.Client
	syncInterval  time.Duration
	syncBatchSize int
}

func NewMeilisearchService() *MeilisearchService {
	s := MeilisearchService{}
	s.init()
	return &s
}

// handleError 统一处理错误并记录日志
func (s *MeilisearchService) handleError(operation string, indexName string, err error) error {
	if err != nil {
		slog.Error(
			"meilisearch: operation failed",
			"operation", operation,
			"indexName", indexName,
			"error", err.Error(),
		)
		return fmt.Errorf("%s failed for index %s: %w", operation, indexName, err)
	}
	return nil
}

// waitForTask 等待任务完成并处理可能的错误
func (s *MeilisearchService) waitForTask(taskUID int64, operation string, indexName string) error {
	task, err := s.client.WaitForTask(taskUID)
	if err != nil {
		err := fmt.Errorf("wait for task %d failed: %w", taskUID, err)
		return s.handleError(operation, indexName, err)
	}

	if task.Status != meilisearch.TaskStatusSucceeded {
		err := fmt.Errorf("wait for task %d failed with error: %s", taskUID, task.Error)
		return s.handleError(operation, indexName, err)
	}

	return nil
}

// Add 将数据添加到MeiliSearch中
func (s *MeilisearchService) Add(indexName string, documents interface{}) error {
	index, err := s.client.GetIndex(indexName)
	if err != nil {
		return s.handleError("get_index", indexName, err)
	}

	// 创建包含评论文本的扩展文档
	documents, err = s.createDocumentsWithComments(indexName, documents)
	if err != nil {
		return s.handleError("create_extended_docs", indexName, err)
	}

	info, err := index.AddDocuments(documents)
	if err != nil {
		return s.handleError("add_documents", indexName, err)
	}

	return s.waitForTask(info.TaskUID, "add_documents", indexName)
}

// Delete 将数据从MeiliSearch中删除
func (s *MeilisearchService) Delete(indexName string, ids []string) error {
	index, err := s.client.GetIndex(indexName)
	if err != nil {
		return s.handleError("get_index", indexName, err)
	}

	info, err := index.DeleteDocuments(ids)
	if err != nil {
		return s.handleError("delete_documents", indexName, err)
	}

	return s.waitForTask(info.TaskUID, "delete_documents", indexName)
}

// Search 搜索
func (s *MeilisearchService) Search(result interface{}, indexName string, query string, page uint, limit uint, sort []string, filter []string) error {
	response, err := s.client.Index(indexName).Search(query, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(page * limit),
		Sort:   sort,
		Filter: filter,
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
		return err
	}
	return nil
}

// createDocumentsWithComments 创建包含评论文本的扩展文档
func (s *MeilisearchService) createDocumentsWithComments(indexName string, objects interface{}) ([]map[string]interface{}, error) {
	// 解码原始对象到map
	data, err := json.Marshal(objects)
	if err != nil {
		return nil, err
	}

	var originalDocs []map[string]interface{}
	if err := json.Unmarshal(data, &originalDocs); err != nil {
		// 如果无法解析为文档数组，尝试解析为单个文档
		var singleDoc map[string]interface{}
		if err := json.Unmarshal(data, &singleDoc); err != nil {
			return nil, err
		}
		originalDocs = []map[string]interface{}{singleDoc}
	}

	// 如果没有文档，直接返回
	if len(originalDocs) == 0 {
		return originalDocs, nil
	}

	// 提取所有ID
	var ids []string
	for _, doc := range originalDocs {
		if id, ok := doc["id"].(float64); ok {
			ids = append(ids, fmt.Sprintf("%s%d", indexName, int(id)))
		}
	}

	// 查询直接评论
	var comments []database.Comment
	if err := database.DB.Where("obj IN ?", ids).Find(&comments).Error; err != nil {
		return nil, err
	}

	// 按对象ID分组评论
	commentsByObjID := make(map[string][]string)
	for _, comment := range comments {
		commentsByObjID[comment.Obj] = append(commentsByObjID[comment.Obj], comment.Text)
	}

	// 构建评论Obj到父Obj的映射 和 评论对象列表用于查询子评论
	superObjMap := make(map[string]string)
	commentObjs := make([]string, len(comments))
	for _, comment := range comments {
		commentObj := fmt.Sprintf("comment%d", comment.ID)
		superObjMap[commentObj] = comment.Obj
		commentObjs = append(commentObjs, commentObj)
	}

	// 查询子评论
	var subComments []database.Comment
	if err := database.DB.Where("obj IN ?", commentObjs).Find(&subComments).Error; err != nil {
		return nil, err
	}

	// 按父评论ID分组子评论
	for _, subComment := range subComments {
		if superObj, ok := superObjMap[subComment.Obj]; ok {
			commentsByObjID[superObj] = append(commentsByObjID[superObj], subComment.Text)
		}
	}

	// 为每个文档添加评论文本字段
	for _, doc := range originalDocs {
		if id, ok := doc["id"].(float64); ok {
			objID := fmt.Sprintf("%s%d", indexName, int(id))
			if comments, ok := commentsByObjID[objID]; ok {
				doc["comments_text"] = strings.Join(comments, " ")
			} else {
				doc["comments_text"] = ""
			}
		}
	}

	return originalDocs, nil
}

// syncDatabase 同步数据库到搜索引擎
func (s *MeilisearchService) syncDatabase(indexName string, value interface{}) error {
	// 从数据库获取数据
	q := database.DB.Model(value)
	if err := q.Find(value).Error; err != nil {
		return s.handleError("db_query", indexName, err)
	}

	// 构建ID映射
	var ids []uint
	if err := database.DB.Model(value).Pluck("id", &ids).Error; err != nil {
		return s.handleError("pluck_ids", indexName, err)
	}
	dbIds := make(map[string]bool)
	for _, id := range ids {
		dbIds[strconv.Itoa(int(id))] = true
	}

	// 获取搜索引擎中的现有数据
	var deleteIds []string
	request := meilisearch.DocumentsQuery{
		Limit:  int64(s.syncBatchSize),
		Offset: 0,
		Fields: []string{"id"},
	}

	// 分批获取并比对ID
	for {
		var res meilisearch.DocumentsResult
		err := s.client.Index(indexName).GetDocuments(&request, &res)
		if err != nil {
			return s.handleError("get_documents", indexName, err)
		}

		if len(res.Results) == 0 {
			break
		}

		for _, v := range res.Results {
			idStr := strconv.Itoa(int(v["id"].(float64)))
			if _, ok := dbIds[idStr]; !ok {
				deleteIds = append(deleteIds, idStr)
			}
		}

		request.Offset += int64(s.syncBatchSize)
	}

	// 添加/更新数据
	if err := s.Add(indexName, value); err != nil {
		return err
	}

	// 删除不存在的数据
	if len(deleteIds) > 0 {
		if err := s.Delete(indexName, deleteIds); err != nil {
			return err
		}
	}

	slog.Info("meilisearch: sync completed", "indexName", indexName)
	return nil
}

// sync 同步数据库到搜索引擎
func (s *MeilisearchService) sync() {
	defer (func() {
		if err := recover(); err != nil {
			slog.Error("meilisearch: sync failed", "err", err)
		} else {
			slog.Info("meilisearch: sync success")
		}
	})()

	s.syncDatabase("course", &[]database.Course{})
	s.syncDatabase("paper", &[]database.Paper{})
	s.syncDatabase("poster", &[]database.Poster{})
}

// initIndex 创建并配置索引
func (s *MeilisearchService) initIndex(indexName, primaryKey string, sortAttr []string, filterAttr []string, searchAttr []string) {
	var indexToUpdate *meilisearch.Index
	// 检查索引是否存在
	indexToUpdate, err := s.client.GetIndex(indexName)
	if err != nil {
		slog.Info("meilisearch: index not found, creating", "indexName", indexName)
		// 创建索引
		index, err := s.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: primaryKey,
		})
		if err != nil {
			panic(s.handleError("create_index", indexName, err))
		}

		// 等待创建完成
		if err := s.waitForTask(index.TaskUID, "create_index", indexName); err != nil {
			panic(err)
		}

		// 获取索引
		indexToUpdate, err = s.client.GetIndex(indexName)
		if err != nil {
			panic(s.handleError("get_index", indexName, err))
		}
	}

	// 更新可排序属性
	index, err := indexToUpdate.UpdateSortableAttributes(&sortAttr)
	if err != nil {
		panic(s.handleError("update_sortable", indexName, err))
	}
	if err := s.waitForTask(index.TaskUID, "update_sortable", indexName); err != nil {
		panic(err)
	}

	// 更新可过滤属性
	index, err = indexToUpdate.UpdateFilterableAttributes(&filterAttr)
	if err != nil {
		panic(s.handleError("update_filterable", indexName, err))
	}
	if err := s.waitForTask(index.TaskUID, "update_filterable", indexName); err != nil {
		panic(err)
	}

	// 更新可搜索属性
	index, err = indexToUpdate.UpdateSearchableAttributes(&searchAttr)
	if err != nil {
		panic(s.handleError("update_searchable", indexName, err))
	}
	if err := s.waitForTask(index.TaskUID, "update_searchable", indexName); err != nil {
		panic(err)
	}

	slog.Info("meilisearch: index initialized", "indexName", indexName)
}

// init 初始化
func (s *MeilisearchService) init() {
	s.client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.Get().Meilisearch.Url,
		APIKey: config.Get().Meilisearch.MasterKey,
	})
	s.syncInterval = time.Duration(config.Get().SyncInterval) * time.Second
	s.syncBatchSize = 1000

	// 可排序参数
	sortAttrCourse := []string{"like_num", "comment_num", "rate", "create_time", "update_time"}
	sortAttrPaper := []string{"like_num", "comment_num", "create_time", "update_time", "edit_time"}
	sortAttrPoster := []string{"like_num", "comment_num", "create_time", "update_time"}
	// 可筛选参数
	filterAttrCourse := []string{}
	filterAttrPaper := []string{}
	filterAttrPoster := []string{"uid", "public", "anonymous"}
	// 可搜索参数
	searchAttrCourse := []string{"name", "number", "teachers_name", "teachers_number", "teachers", "comments_text"}
	searchAttrPaper := []string{"title", "intro", "content", "comments_text"}
	searchAttrPoster := []string{"title", "text", "tags", "comments_text"}

	s.initIndex("course", "id", sortAttrCourse, filterAttrCourse, searchAttrCourse)
	s.initIndex("paper", "id", sortAttrPaper, filterAttrPaper, searchAttrPaper)
	s.initIndex("poster", "id", sortAttrPoster, filterAttrPoster, searchAttrPoster)

	// 与数据库同步
	go func() {
		for {
			go s.sync()
			time.Sleep(s.syncInterval)
		}
	}()
}
