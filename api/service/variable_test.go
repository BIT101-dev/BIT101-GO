package service

import (
	"BIT101-GO/database"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 设置模拟数据库
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建sqlmock数据库失败: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: db,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("打开gorm数据库失败: %v", err)
	}

	database.DB = gormDB

	return db, mock, func() {
		db.Close()
	}
}

func TestVariableService_Get_Success(t *testing.T) {
	_, mock, cleanup := setupMockDB(t)
	defer cleanup()

	service := NewVariableService()
	objID := "test_obj"
	expectedData := "test_data"

	rows := sqlmock.NewRows([]string{"id", "obj", "data"}).
		AddRow(1, objID, expectedData)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"variables\" WHERE obj = ? LIMIT 1")).
		WithArgs(objID).
		WillReturnRows(rows)

	data, err := service.Get(objID)
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
	}
}

func TestVariableService_Get_NotFound(t *testing.T) {
	_, mock, cleanup := setupMockDB(t)
	defer cleanup()

	service := NewVariableService()
	objID := "nonexistent_obj"

	rows := sqlmock.NewRows([]string{"id", "obj", "data"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `variables` WHERE obj = ? LIMIT 1")).
		WithArgs(objID).
		WillReturnRows(rows)

	_, err := service.Get(objID)
	assert.Error(t, err)
	assert.Equal(t, "变量不存在Orz", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
	}
}

func TestVariableService_Get_DatabaseError(t *testing.T) {
	_, mock, cleanup := setupMockDB(t)
	defer cleanup()

	service := NewVariableService()
	objID := "test_obj"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `variables` WHERE obj = ? LIMIT 1")).
		WithArgs(objID).
		WillReturnError(sql.ErrConnDone)

	_, err := service.Get(objID)
	assert.Error(t, err)
	assert.Equal(t, "数据库错误Orz", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
	}
}

func TestVariableService_Set_Success(t *testing.T) {
	_, mock, cleanup := setupMockDB(t)
	defer cleanup()

	service := NewVariableService()
	objID := "test_obj"
	data := "updated_data"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `variables` SET `data` = ? WHERE obj = ?")).
		WithArgs(data, objID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.Set(objID, data)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
	}
}

func TestVariableService_Set_DatabaseError(t *testing.T) {
	_, mock, cleanup := setupMockDB(t)
	defer cleanup()

	service := NewVariableService()
	objID := "test_obj"
	data := "updated_data"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `variables` SET `data` = ? WHERE obj = ?")).
		WithArgs(data, objID).
		WillReturnError(sql.ErrConnDone)

	err := service.Set(objID, data)
	assert.Error(t, err)
	assert.Equal(t, "数据库错误Orz", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
	}
}

func TestNewVariableService(t *testing.T) {
	service := NewVariableService()
	assert.NotNil(t, service)
	assert.IsType(t, &VariableService{}, service)
}
