package tests

import (
	"context"
	"sync"
	"testing"

	orderModel "order-management/order/model"
	orderService "order-management/order/service"
	productModel "order-management/product/model"
)

type mockOrderRepo struct {
	sync.Mutex
	orders map[uint]*orderModel.Order
	count  uint
}

func (m *mockOrderRepo) Create(ctx context.Context, o *orderModel.Order) error {
	m.Lock()
	defer m.Unlock()
	m.count++
	o.ID = m.count
	m.orders[o.ID] = o
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uint) (*orderModel.Order, error) {
	m.Lock()
	defer m.Unlock()
	return m.orders[id], nil
}

func (m *mockOrderRepo) List(ctx context.Context, limit, offset int) ([]orderModel.Order, error) {
	m.Lock()
	defer m.Unlock()
	res := make([]orderModel.Order, 0, len(m.orders))
	for _, v := range m.orders {
		res = append(res, *v)
	}
	return res, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uint, status orderModel.Status) error {
	m.Lock()
	defer m.Unlock()
	if o, ok := m.orders[id]; ok {
		o.Status = status
	}
	return nil
}

func (m *mockOrderRepo) Ping(ctx context.Context) error { return nil }

type mockProductService struct{}

func (m *mockProductService) GetProductPrice(ctx context.Context, id uint) (float64, error) {
	return 25.50, nil
}

func (m *mockProductService) GetProductDetails(ctx context.Context, id uint) (*productModel.Product, error) {
	return &productModel.Product{ID: id, Price: 25.50, Stock: 100}, nil
}

func TestConcurrentOrders(t *testing.T) {
	repo := &mockOrderRepo{orders: make(map[uint]*orderModel.Order)}
	prodSvc := &mockProductService{}
	svc := orderService.NewService(repo, prodSvc)

	const concurrencyCount = 100
	var wg sync.WaitGroup
	wg.Add(concurrencyCount)

	for i := 0; i < concurrencyCount; i++ {
		go func(idx int) {
			defer wg.Done()
			req := &orderModel.CreateRequest{
				CustomerID: uint(100 + idx),
				ProductID:  uint(500 + idx),
				Quantity:   2,
			}
			res, err := svc.CreateOrder(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if res == nil || res.ID == 0 {
				t.Errorf("Order failed to initialize")
			}
		}(i)
	}

	wg.Wait()

	if repo.count != concurrencyCount {
		t.Fatalf("Expected %d orders, got %d", concurrencyCount, repo.count)
	}
}