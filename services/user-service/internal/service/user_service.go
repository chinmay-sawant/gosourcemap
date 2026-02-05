package service

import (
	"context"
	"sync"

	"user-service/internal/client"
	"user-service/internal/models"
)

// UserService defines business logic for users.
type UserService interface {
	GetAll() []models.User
	GetByID(id string) (*models.User, error)
	Create(req models.CreateUserRequest) *models.User
	GetDashboard(ctx context.Context, userID string) (*models.UserDashboard, error)
}

type userService struct {
	users           map[string]models.User
	orderClient     *client.HTTPClient
	inventoryClient *client.HTTPClient
}

// NewUserService creates a new UserService.
func NewUserService(orderClient, inventoryClient *client.HTTPClient) UserService {
	// Seed with sample data
	users := map[string]models.User{
		"user-1": {ID: "user-1", Name: "Alice Johnson", Email: "alice@example.com"},
		"user-2": {ID: "user-2", Name: "Bob Smith", Email: "bob@example.com"},
	}
	return &userService{
		users:           users,
		orderClient:     orderClient,
		inventoryClient: inventoryClient,
	}
}

func (s *userService) GetAll() []models.User {
	result := make([]models.User, 0, len(s.users))
	for _, u := range s.users {
		result = append(result, u)
	}
	return result
}

func (s *userService) GetByID(id string) (*models.User, error) {
	if u, ok := s.users[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (s *userService) Create(req models.CreateUserRequest) *models.User {
	id := "user-new"
	user := models.User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[id] = user
	return &user
}

// GetDashboard orchestrates calls to Order and Inventory services.
func (s *userService) GetDashboard(ctx context.Context, userID string) (*models.UserDashboard, error) {
	user, _ := s.GetByID(userID)
	if user == nil {
		return nil, nil
	}

	var wg sync.WaitGroup
	var orders, inventory interface{}
	var orderErr, inventoryErr error

	// Parallel calls to Order and Inventory services
	wg.Add(2)

	go func() {
		defer wg.Done()
		var result map[string]interface{}
		orderErr = s.orderClient.Get(ctx, "/v1/orders?user_id="+userID, &result)
		if orderErr == nil {
			orders = result
		}
	}()

	go func() {
		defer wg.Done()
		var result map[string]interface{}
		inventoryErr = s.inventoryClient.Get(ctx, "/v1/inventory", &result)
		if inventoryErr == nil {
			inventory = result
		}
	}()

	wg.Wait()

	// Return partial data even if some calls fail
	dashboard := &models.UserDashboard{
		User:      *user,
		Orders:    orders,
		Inventory: inventory,
	}

	return dashboard, nil
}
