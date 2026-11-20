package todo

import (
	"app/internal/database"
	"gorm.io/gorm"
)

type TodoRepository interface {
	// Todos
	FindTodosByUserID(userID uint, groupID *uint) ([]database.Todo, error)
	FindTodoByID(id uint) (*database.Todo, error)
	CreateTodo(todo *database.Todo) error
	UpdateTodo(todo *database.Todo) error
	DeleteTodo(id uint) error
	GetStatsByUserID(userID uint) (int64, int64, int64, error)

	// Groups
	FindGroupsByUserID(userID uint) ([]database.TodoGroup, error)
	FindGroupByID(id uint) (*database.TodoGroup, error)
	CreateGroup(group *database.TodoGroup) error
	UpdateGroup(group *database.TodoGroup) error
	DeleteGroup(id uint) error

	// Templates (Public)
	FindPublicTemplates() ([]database.TodoGroup, error)
	FindPublicTemplateByID(id uint) (*database.TodoGroup, error)
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) FindTodosByUserID(userID uint, groupID *uint) ([]database.Todo, error) {
	var todos []database.Todo
	query := r.db.Where("user_id = ?", userID)
	if groupID != nil && *groupID > 0 {
		query = query.Where("todo_group_id = ?", *groupID)
	}
	err := query.Order("created_at desc").Find(&todos).Error
	return todos, err
}

func (r *todoRepository) FindTodoByID(id uint) (*database.Todo, error) {
	var todo database.Todo
	err := r.db.First(&todo, id).Error
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) CreateTodo(todo *database.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) UpdateTodo(todo *database.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) DeleteTodo(id uint) error {
	return r.db.Delete(&database.Todo{}, id).Error
}

func (r *todoRepository) GetStatsByUserID(userID uint) (int64, int64, int64, error) {
	var total, selesai, belumSelesai int64
	r.db.Model(&database.Todo{}).Where("user_id = ?", userID).Count(&total)
	r.db.Model(&database.Todo{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&selesai)
	r.db.Model(&database.Todo{}).Where("user_id = ? AND status != ?", userID, "completed").Count(&belumSelesai)
	return total, selesai, belumSelesai, nil
}

func (r *todoRepository) FindGroupsByUserID(userID uint) ([]database.TodoGroup, error) {
	var groups []database.TodoGroup
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&groups).Error
	return groups, err
}

func (r *todoRepository) FindGroupByID(id uint) (*database.TodoGroup, error) {
	var group database.TodoGroup
	err := r.db.Preload("Todos").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *todoRepository) CreateGroup(group *database.TodoGroup) error {
	return r.db.Create(group).Error
}

func (r *todoRepository) UpdateGroup(group *database.TodoGroup) error {
	return r.db.Save(group).Error
}

func (r *todoRepository) DeleteGroup(id uint) error {
	// Nullify todos belonging to this group
	r.db.Model(&database.Todo{}).Where("todo_group_id = ?", id).Update("todo_group_id", nil)
	return r.db.Delete(&database.TodoGroup{}, id).Error
}

func (r *todoRepository) FindPublicTemplates() ([]database.TodoGroup, error) {
	var groups []database.TodoGroup
	err := r.db.Where("is_public = ?", true).Preload("User").Preload("Todos").Order("created_at desc").Find(&groups).Error
	return groups, err
}

func (r *todoRepository) FindPublicTemplateByID(id uint) (*database.TodoGroup, error) {
	var group database.TodoGroup
	err := r.db.Where("id = ? AND is_public = ?", id, true).Preload("User").Preload("Todos").First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}
