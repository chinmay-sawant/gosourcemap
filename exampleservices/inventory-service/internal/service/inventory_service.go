package service

import (
	"inventory-service/internal/models"
)

// InventoryService defines business logic for inventory.
type InventoryService interface {
	GetAll() []models.InventoryItem
	GetByID(id string) (*models.InventoryItem, error)
	Update(id string, req models.UpdateInventoryRequest) (*models.InventoryItem, error)
}

type inventoryService struct {
	items map[string]models.InventoryItem
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService() InventoryService {
	// Seed with sample data
	items := map[string]models.InventoryItem{
		"prod-101": {ID: "prod-101", ProductName: "Wireless Mouse", SKU: "WM-001", Quantity: 150, Price: 29.99},
		"prod-102": {ID: "prod-102", ProductName: "Mechanical Keyboard", SKU: "MK-001", Quantity: 75, Price: 149.99},
		"prod-103": {ID: "prod-103", ProductName: "USB-C Hub", SKU: "UH-001", Quantity: 200, Price: 49.99},
	}
	return &inventoryService{items: items}
}

func (s *inventoryService) GetAll() []models.InventoryItem {
	result := make([]models.InventoryItem, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *inventoryService) GetByID(id string) (*models.InventoryItem, error) {
	if item, ok := s.items[id]; ok {
		return &item, nil
	}
	return nil, nil
}

func (s *inventoryService) Update(id string, req models.UpdateInventoryRequest) (*models.InventoryItem, error) {
	if item, ok := s.items[id]; ok {
		item.Quantity = req.Quantity
		s.items[id] = item
		return &item, nil
	}
	return nil, nil
}
