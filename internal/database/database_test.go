package database

import (
	"testing"

	"app/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConnectTest_And_Migrate_And_Seed(t *testing.T) {
	db, err := ConnectTest()
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test Seeding with config
	cfg := &config.Config{
		AdminUsername: "admin",
		AdminPassword: "admin",
	}
	err = Seed(db, cfg)
	assert.NoError(t, err)

	// Verify Admin
	var admin User
	err = db.Where("username = ?", "admin").First(&admin).Error
	assert.NoError(t, err)
	assert.Equal(t, "admin", admin.Role)

	// Verify Member
	var member User
	err = db.Where("username = ?", "member").First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, "member", member.Role)

	// Calling Seed again without config should be idempotent
	err = Seed(db)
	assert.NoError(t, err)

	var userCount int64
	db.Model(&User{}).Count(&userCount)
	assert.Equal(t, int64(2), userCount)

	// Verify 100 default todos (50 template todos, 50 personal todos)
	var totalTodos int64
	db.Model(&Todo{}).Count(&totalTodos)
	assert.Equal(t, int64(100), totalTodos)

	var templateTodos int64
	db.Model(&Todo{}).Where("todo_group_id IS NOT NULL").Count(&templateTodos)
	assert.Equal(t, int64(50), templateTodos)

	var personalTodos int64
	db.Model(&Todo{}).Where("todo_group_id IS NULL").Count(&personalTodos)
	assert.Equal(t, int64(50), personalTodos)
}

func TestConnect_Sqlite(t *testing.T) {
	cfg := &config.Config{
		DBConnection: "sqlite",
		DBDatabase:   ":memory:",
	}
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = Migrate(db)
	assert.NoError(t, err)
}

func TestConnect_Postgres_Invalid(t *testing.T) {
	// Connecting to non-existent postgres port should return error or fail gracefully
	cfg := &config.Config{
		DBConnection: "pgsql",
		DBHost:       "127.0.0.1",
		DBPort:       "54399",
		DBDatabase:   "invalid_db",
		DBUsername:   "invalid_user",
		DBPassword:   "invalid_pass",
		DBSSLMode:    "disable",
	}
	db, err := Connect(cfg)
	if err == nil && db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}
