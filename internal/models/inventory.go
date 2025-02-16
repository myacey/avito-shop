package models

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int32  `json:"quantity"`
}
