package models

type User struct {
	ID           int32            `json:"-"`
	Username     string           `json:"-"`
	Password     string           `json:"-"`
	Coins        int32            `json:"coins"`
	Inventory    []*InventoryItem `json:"inventory"`
	EntryHistory interface{}      `json:"coinHistory"`
}
