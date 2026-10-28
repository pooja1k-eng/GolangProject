package repository

import (
	"context"

	productModel "order-management/product/model"

	"gorm.io/gorm"
)

type Repository interface {
	GetByID(ctx context.Context, id uint) (*productModel.Product, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uint) (*productModel.Product, error) {
	var p productModel.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}