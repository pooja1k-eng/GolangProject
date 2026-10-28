package repository

import (
	"context"

	orderModel "order-management/order/model"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, order *orderModel.Order) error
	GetByID(ctx context.Context, id uint) (*orderModel.Order, error)
	List(ctx context.Context, limit, offset int) ([]orderModel.Order, error)
	UpdateStatus(ctx context.Context, id uint, status orderModel.Status) error
	Ping(ctx context.Context) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, o *orderModel.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*orderModel.Order, error) {
	var o orderModel.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]orderModel.Order, error) {
	var orders []orderModel.Order
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id desc").Find(&orders).Error
	return orders, err
}

func (r *repository) UpdateStatus(ctx context.Context, id uint, status orderModel.Status) error {
	res := r.db.WithContext(ctx).Model(&orderModel.Order{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}