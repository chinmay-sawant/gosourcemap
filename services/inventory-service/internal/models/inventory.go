package models

// InventoryItem represents an inventory entity.
type InventoryItem struct {
	ID          string  `json:"id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// UpdateInventoryRequest is the payload to update inventory.
type UpdateInventoryRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}
