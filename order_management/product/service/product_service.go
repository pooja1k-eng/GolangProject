package service

import (
	"context"

	productModel "order-management/product/model"
	productRepo "order-management/product/repository"
)

type Service interface {
	GetProductPrice(ctx context.Context, productID uint) (float64, error)
	GetProductDetails(ctx context.Context, productID uint) (*productModel.Product, error)
}

type service struct {
	repo productRepo.Repository
}

func NewService(repo productRepo.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProductPrice(ctx context.Context, productID uint) (float64, error) {
	prod, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return 0, err
	}
	return prod.Price, nil
}

func (s *service) GetProductDetails(ctx context.Context, productID uint) (*productModel.Product, error) {
	return s.repo.GetByID(ctx, productID)
}