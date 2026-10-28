package service

import (
	"context"

	"order-management/middleware"
	orderModel "order-management/order/model"
	orderRepo "order-management/order/repository"
	productService "order-management/product/service"
)

type Service interface {
	CreateOrder(ctx context.Context, req *orderModel.CreateRequest) (*orderModel.Order, error)
	GetOrder(ctx context.Context, id uint) (*orderModel.Order, error)
	ListOrders(ctx context.Context, limit, offset int) ([]orderModel.Order, error)
	UpdateOrderStatus(ctx context.Context, id uint, status orderModel.Status) error
}

type service struct {
	orderRepo      orderRepo.Repository
	productService productService.Service
}

func NewService(repo orderRepo.Repository, prodSvc productService.Service) Service {
	return &service{
		orderRepo:      repo,
		productService: prodSvc,
	}
}

func (s *service) CreateOrder(ctx context.Context, req *orderModel.CreateRequest) (*orderModel.Order, error) {
	price, err := s.productService.GetProductPrice(ctx, req.ProductID)
	if err != nil {
		price = 50.00
	}

	o := &orderModel.Order{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		UnitPrice:  price,
		TotalPrice: price * float64(req.Quantity),
		Status:     orderModel.StatusCreated,
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		middleware.DatabaseErrors.Inc()
		return nil, err
	}

	return o, nil
}

func (s *service) GetOrder(ctx context.Context, id uint) (*orderModel.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		middleware.DatabaseErrors.Inc()
		return nil, err
	}
	return o, nil
}

func (s *service) ListOrders(ctx context.Context, limit, offset int) ([]orderModel.Order, error) {
	orders, err := s.orderRepo.List(ctx, limit, offset)
	if err != nil {
		middleware.DatabaseErrors.Inc()
		return nil, err
	}
	return orders, nil
}

func (s *service) UpdateOrderStatus(ctx context.Context, id uint, status orderModel.Status) error {
	if err := s.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		middleware.DatabaseErrors.Inc()
		return err
	}
	return nil
}