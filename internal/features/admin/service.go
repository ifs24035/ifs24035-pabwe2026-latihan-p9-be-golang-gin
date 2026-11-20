package admin

import (
	"errors"
	"time"

	"app/internal/database"
	"gorm.io/gorm"
)

type UserWithStats struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	TotalTodos   int64     `json:"total_todos"`
	TotalSelesai int64     `json:"total_selesai"`
}

type AdminService interface {
	GetAllUsers() ([]UserWithStats, error)
	GetUserTodos(targetUserID uint) ([]database.Todo, error)
}

type adminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) AdminService {
	return &adminService{db: db}
}

func (s *adminService) GetAllUsers() ([]UserWithStats, error) {
	var users []database.User
	_ = s.db.Find(&users).Error

	var results []UserWithStats
	for _, u := range users {
		var total, selesai int64
		s.db.Model(&database.Todo{}).Where("user_id = ?", u.ID).Count(&total)
		s.db.Model(&database.Todo{}).Where("user_id = ? AND status = ?", u.ID, "completed").Count(&selesai)

		results = append(results, UserWithStats{
			ID:           u.ID,
			Username:     u.Username,
			Email:        u.Email,
			Role:         u.Role,
			CreatedAt:    u.CreatedAt,
			TotalTodos:   total,
			TotalSelesai: selesai,
		})
	}
	return results, nil
}

func (s *adminService) GetUserTodos(targetUserID uint) ([]database.Todo, error) {
	var user database.User
	if err := s.db.First(&user, targetUserID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var todos []database.Todo
	err := s.db.Where("user_id = ?", targetUserID).Order("created_at desc").Find(&todos).Error
	return todos, err
}
