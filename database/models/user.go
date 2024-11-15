package models

import "time"

type User struct {
	ID        uint32    `gorm:"primaryKey"`
	Nickname  string    `gorm:"unique"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
