package models

// User represents a user entity.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUserRequest is the payload to create a user.
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UserDashboard aggregates data from multiple services.
type UserDashboard struct {
	User      User        `json:"user"`
	Orders    interface{} `json:"orders"`
	Inventory interface{} `json:"inventory"`
}
