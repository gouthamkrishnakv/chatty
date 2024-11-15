package models

import (
	"time"
)

type Message struct {
	ID        uint32    `gorm:"primaryKey"`
	AuthorID  uint32    `gorm:"not null"`
	Author    User      `gorm:"foreignKey:AuthorID"`
	Message   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
