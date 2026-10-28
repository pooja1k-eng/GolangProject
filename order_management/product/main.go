package main

import (
	"context"
	"errors"
	"fmt"
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
	productModel "order-management/product/model"
	productRepo "order-management/product/repository"
	productService "order-management/product/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	port := cfg.ProductServicePort

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("[Product Service] Database init failed: %v", err)
	}

	seedProducts(db)

	repo := productRepo.NewRepository(db)
	svc := productService.NewService(repo)

	router := gin.New()
	router.Use(
		middleware.SecurityHeaders(),
		middleware.RequestID(),
		middleware.StructuredLogger(),
		middleware.Recovery(),
	)

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "product-service", "status": "UP"})
	})

	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		p, err := svc.GetProductDetails(c.Request.Context(), uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
			return
		}

		c.JSON(http.StatusOK, p)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("=====================================================")
		log.Printf("  Product Service running on http://localhost:%s", port)
		log.Printf("=====================================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Product Service] Listener error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("[Product Service] Terminated cleanly.")
}

func seedProducts(db *gorm.DB) {
	sampleProducts := []productModel.Product{
		{ID: 501, Name: "Mechanical Keyboard", Price: 89.99, Stock: 50},
		{ID: 502, Name: "Wireless Mouse", Price: 49.99, Stock: 120},
		{ID: 503, Name: "USB-C Monitor Hub", Price: 129.50, Stock: 35},
	}
	for _, p := range sampleProducts {
		db.FirstOrCreate(&p, productModel.Product{ID: p.ID})
	}
	fmt.Println("[Product Service] Products seeded successfully (IDs: 501, 502, 503).")
}