package model

type Category struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}
