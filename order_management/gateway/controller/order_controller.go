package controller

import (
	"errors"
	"net/http"
	"strconv"

	orderModel "order-management/order/model"
	orderRepo "order-management/order/repository"
	orderService "order-management/order/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct {
	orderService orderService.Service
	orderRepo    orderRepo.Repository
}

func NewOrderController(svc orderService.Service, repo orderRepo.Repository) *OrderController {
	return &OrderController{orderService: svc, orderRepo: repo}
}

func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var req orderModel.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	res, err := ctrl.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (ctrl *OrderController) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order id format"})
		return
	}

	res, err := ctrl.orderService.GetOrder(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *OrderController) ListOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	res, err := ctrl.orderService.ListOrders(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *OrderController) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order id format"})
		return
	}

	var req orderModel.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	if err := ctrl.orderService.UpdateOrderStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (ctrl *OrderController) HealthLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func (ctrl *OrderController) HealthReady(c *gin.Context) {
	if err := ctrl.orderRepo.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN", "error": "Database unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "READY"})
}