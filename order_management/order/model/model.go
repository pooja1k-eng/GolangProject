package model

import "time"

type Status string

const (
	StatusCreated   Status = "CREATED"
	StatusConfirmed Status = "CONFIRMED"
	StatusCancelled Status = "CANCELLED"
)

type Order struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CustomerID uint      `gorm:"index;not null" json:"customerId"`
	ProductID  uint      `gorm:"index;not null" json:"productId"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	UnitPrice  float64   `gorm:"type:numeric(10,2);not null" json:"unitPrice"`
	TotalPrice float64   `gorm:"type:numeric(10,2);not null" json:"totalPrice"`
	Status     Status    `gorm:"type:varchar(20);index;not null" json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CreateRequest struct {
	CustomerID uint `json:"customerId" binding:"required,gt=0"`
	ProductID  uint `json:"productId" binding:"required,gt=0"`
	Quantity   int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateStatusRequest struct {
	Status Status `json:"status" binding:"required,oneof=CREATED CONFIRMED CANCELLED"`
}