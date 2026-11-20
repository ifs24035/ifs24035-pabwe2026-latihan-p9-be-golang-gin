package auth

import (
	"app/internal/database"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByUsernameOrEmail(identifier string) (*database.User, error)
	FindByID(id uint) (*database.User, error)
	Create(user *database.User) error
	Update(user *database.User) error
	FindAll() ([]database.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByUsernameOrEmail(identifier string) (*database.User, error) {
	var user database.User
	err := r.db.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*database.User, error) {
	var user database.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *database.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *database.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) FindAll() ([]database.User, error) {
	var users []database.User
	err := r.db.Find(&users).Error
	return users, err
}
