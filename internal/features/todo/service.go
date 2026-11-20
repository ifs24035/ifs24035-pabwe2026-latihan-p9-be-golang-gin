package todo

import (
	"errors"
	"fmt"

	"app/internal/database"
)

type TodoService interface {
	// Todos
	GetTodos(userID uint, groupID *uint) ([]database.Todo, error)
	GetTodoByID(userID uint, id uint) (*database.Todo, error)
	CreateTodo(userID uint, req CreateTodoRequest) (*database.Todo, error)
	UpdateTodo(userID uint, id uint, req UpdateTodoRequest) (*database.Todo, error)
	UpdateTodoStatus(userID uint, id uint, status string) (*database.Todo, error)
	DeleteTodo(userID uint, id uint) error
	GetStats(userID uint) (*TodoStatsResponse, error)

	// Groups
	GetGroups(userID uint) ([]database.TodoGroup, error)
	GetGroupByID(userID uint, id uint) (*database.TodoGroup, error)
	CreateGroup(userID uint, req CreateTodoGroupRequest) (*database.TodoGroup, error)
	UpdateGroup(userID uint, id uint, req UpdateTodoGroupRequest) (*database.TodoGroup, error)
	DeleteGroup(userID uint, id uint) error

	// Templates & Clone
	GetPublicTemplates() ([]database.TodoGroup, error)
	GetPublicTemplateByID(id uint) (*database.TodoGroup, error)
	CloneTemplate(userID uint, templateID uint) (*database.TodoGroup, error)
}

type todoService struct {
	repo TodoRepository
}

func NewTodoService(repo TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) GetTodos(userID uint, groupID *uint) ([]database.Todo, error) {
	return s.repo.FindTodosByUserID(userID, groupID)
}

func (s *todoService) GetTodoByID(userID uint, id uint) (*database.Todo, error) {
	todo, err := s.repo.FindTodoByID(id)
	if err != nil {
		return nil, errors.New("todo not found")
	}
	if todo.UserID != userID {
		return nil, errors.New("forbidden: you do not own this todo")
	}
	return todo, nil
}

func (s *todoService) CreateTodo(userID uint, req CreateTodoRequest) (*database.Todo, error) {
	status := req.Status
	if status == "" {
		status = "pending"
	}
	todo := &database.Todo{
		UserID:         userID,
		TodoGroupID:    req.TodoGroupID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         status,
		DueDate:        req.DueDate,
		CoverImagePath: req.CoverImagePath,
	}
	_ = s.repo.CreateTodo(todo)
	return todo, nil
}

func (s *todoService) UpdateTodo(userID uint, id uint, req UpdateTodoRequest) (*database.Todo, error) {
	todo, err := s.GetTodoByID(userID, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	if req.Status != "" {
		todo.Status = req.Status
	}
	if req.DueDate != nil {
		todo.DueDate = req.DueDate
	}
	if req.CoverImagePath != "" {
		if req.CoverImagePath == "DELETE" || req.CoverImagePath == "remove" {
			todo.CoverImagePath = ""
		} else {
			todo.CoverImagePath = req.CoverImagePath
		}
	}
	if req.TodoGroupID != nil {
		todo.TodoGroupID = req.TodoGroupID
	}

	_ = s.repo.UpdateTodo(todo)
	return todo, nil
}

func (s *todoService) UpdateTodoStatus(userID uint, id uint, status string) (*database.Todo, error) {
	todo, err := s.GetTodoByID(userID, id)
	if err != nil {
		return nil, err
	}
	todo.Status = status
	_ = s.repo.UpdateTodo(todo)
	return todo, nil
}

func (s *todoService) DeleteTodo(userID uint, id uint) error {
	_, err := s.GetTodoByID(userID, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteTodo(id)
}

func (s *todoService) GetStats(userID uint) (*TodoStatsResponse, error) {
	total, selesai, belumSelesai, _ := s.repo.GetStatsByUserID(userID)
	return &TodoStatsResponse{
		Total:        total,
		TotalSelesai: selesai,
		BelumSelesai: belumSelesai,
	}, nil
}

func (s *todoService) GetGroups(userID uint) ([]database.TodoGroup, error) {
	return s.repo.FindGroupsByUserID(userID)
}

func (s *todoService) GetGroupByID(userID uint, id uint) (*database.TodoGroup, error) {
	group, err := s.repo.FindGroupByID(id)
	if err != nil {
		return nil, errors.New("group not found")
	}
	if group.UserID != userID && !group.IsPublic {
		return nil, errors.New("forbidden: you do not own this group")
	}
	return group, nil
}

func (s *todoService) CreateGroup(userID uint, req CreateTodoGroupRequest) (*database.TodoGroup, error) {
	group := &database.TodoGroup{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}
	_ = s.repo.CreateGroup(group)
	return group, nil
}

func (s *todoService) UpdateGroup(userID uint, id uint, req UpdateTodoGroupRequest) (*database.TodoGroup, error) {
	group, err := s.repo.FindGroupByID(id)
	if err != nil {
		return nil, errors.New("group not found")
	}
	if group.UserID != userID {
		return nil, errors.New("forbidden: you do not own this group")
	}

	if req.Title != "" {
		group.Title = req.Title
	}
	if req.Description != "" {
		group.Description = req.Description
	}
	if req.IsPublic != nil {
		group.IsPublic = *req.IsPublic
	}

	_ = s.repo.UpdateGroup(group)
	return group, nil
}

func (s *todoService) DeleteGroup(userID uint, id uint) error {
	group, err := s.repo.FindGroupByID(id)
	if err != nil {
		return errors.New("group not found")
	}
	if group.UserID != userID {
		return errors.New("forbidden: you do not own this group")
	}
	return s.repo.DeleteGroup(id)
}

func (s *todoService) GetPublicTemplates() ([]database.TodoGroup, error) {
	return s.repo.FindPublicTemplates()
}

func (s *todoService) GetPublicTemplateByID(id uint) (*database.TodoGroup, error) {
	tpl, err := s.repo.FindPublicTemplateByID(id)
	if err != nil {
		return nil, errors.New("template not found")
	}
	return tpl, nil
}

func (s *todoService) CloneTemplate(userID uint, templateID uint) (*database.TodoGroup, error) {
	tpl, err := s.GetPublicTemplateByID(templateID)
	if err != nil {
		return nil, err
	}

	newGroup := &database.TodoGroup{
		UserID:      userID,
		Title:       fmt.Sprintf("%s (Salinan)", tpl.Title),
		Description: tpl.Description,
		IsPublic:    false,
	}
	_ = s.repo.CreateGroup(newGroup)

	for _, item := range tpl.Todos {
		clonedTodo := database.Todo{
			UserID:         userID,
			TodoGroupID:    &newGroup.ID,
			Title:          item.Title,
			Description:    item.Description,
			Status:         "pending",
			DueDate:        item.DueDate,
			CoverImagePath: item.CoverImagePath,
		}
		_ = s.repo.CreateTodo(&clonedTodo)
	}

	return s.repo.FindGroupByID(newGroup.ID)
}
