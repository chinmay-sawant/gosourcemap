package service

import (
	"order-service/internal/models"

	"github.com/google/uuid"
)

// OrderService defines business logic for orders.
type OrderService interface {
	GetAll() []models.Order
	GetByID(id string) (*models.Order, error)
	Create(req models.CreateOrderRequest) *models.Order
	GetByUserID(userID string) []models.Order
}

type orderService struct {
	orders map[string]models.Order
}

// NewOrderService creates a new OrderService.
func NewOrderService() OrderService {
	// Seed with sample data
	orders := map[string]models.Order{
		"ord-001": {ID: "ord-001", UserID: "user-1", ProductID: "prod-101", Quantity: 2, Total: 59.98, Status: "completed"},
		"ord-002": {ID: "ord-002", UserID: "user-1", ProductID: "prod-102", Quantity: 1, Total: 149.99, Status: "pending"},
		"ord-003": {ID: "ord-003", UserID: "user-2", ProductID: "prod-101", Quantity: 5, Total: 149.95, Status: "shipped"},
	}
	return &orderService{orders: orders}
}

func (s *orderService) GetAll() []models.Order {
	result := make([]models.Order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, o)
	}
	return result
}

func (s *orderService) GetByID(id string) (*models.Order, error) {
	if o, ok := s.orders[id]; ok {
		return &o, nil
	}
	return nil, nil
}

func (s *orderService) Create(req models.CreateOrderRequest) *models.Order {
	id := "ord-" + uuid.New().String()[:8]
	order := models.Order{
		ID:        id,
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Total:     float64(req.Quantity) * 29.99, // placeholder price
		Status:    "pending",
	}
	s.orders[id] = order
	return &order
}

func (s *orderService) GetByUserID(userID string) []models.Order {
	result := []models.Order{}
	for _, o := range s.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result
}
