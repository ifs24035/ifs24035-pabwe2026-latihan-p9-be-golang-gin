package todo

import "time"

type CreateTodoRequest struct {
	TodoGroupID    *uint      `json:"todo_group_id"`
	Title          string     `json:"title" binding:"required"`
	Description    string     `json:"description"`
	Status         string     `json:"status"` // pending, in_progress, completed
	DueDate        *time.Time `json:"due_date"`
	CoverImagePath string     `json:"cover_image_path"`
}

type UpdateTodoRequest struct {
	TodoGroupID    *uint      `json:"todo_group_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	DueDate        *time.Time `json:"due_date"`
	CoverImagePath string     `json:"cover_image_path"`
}

type UpdateTodoStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type CreateTodoGroupRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type UpdateTodoGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
}

type TodoStatsResponse struct {
	Total        int64 `json:"total"`
	TotalSelesai int64 `json:"total_selesai"`
	BelumSelesai int64 `json:"belum_selesai"`
}
