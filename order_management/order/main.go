package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"order-management/config"
	"order-management/database"
	"order-management/middleware"
	orderModel "order-management/order/model"
	orderRepo "order-management/order/repository"
	orderService "order-management/order/service"
	productClient "order-management/product/client"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	port := cfg.OrderServicePort // 10004

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("[Order Service] Database init failed: %v", err)
	}

	// Resilient client connected to Product Service at 8081
	remoteProdClient := productClient.NewRemoteProductService(cfg.ProductServiceURL)

	repo := orderRepo.NewRepository(db)
	svc := orderService.NewService(repo, remoteProdClient)

	router := gin.New()
	router.Use(
		middleware.SecurityHeaders(),
		middleware.RequestID(),
		middleware.StructuredLogger(),
		middleware.Recovery(),
	)

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "order-service", "status": "UP"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		if err := repo.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	api := router.Group("/api/v1/orders")
	{
		api.POST("", func(c *gin.Context) {
			var req orderModel.CreateRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := svc.CreateOrder(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
				return
			}
			c.JSON(http.StatusCreated, res)
		})

		api.GET("/:id", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
				return
			}
			res, err := svc.GetOrder(c.Request.Context(), uint(id))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		api.GET("", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
			orders, err := svc.ListOrders(c.Request.Context(), limit, offset)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
				return
			}
			c.JSON(http.StatusOK, orders)
		})

		api.PUT("/:id/status", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
				return
			}
			var req orderModel.UpdateStatusRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.UpdateOrderStatus(c.Request.Context(), uint(id), req.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
		})
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("=====================================================")
		log.Printf("  Order Service running on http://localhost:%s", port)
		log.Printf("=====================================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Order Service] Listener error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("[Order Service] Terminated cleanly.")
}