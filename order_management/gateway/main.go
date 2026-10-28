package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-management/config"
	"order-management/database"
	gatewayGraphQL "order-management/gateway/graphql"
	"order-management/middleware"
	orderRepo "order-management/order/repository"
	orderService "order-management/order/service"
	productClient "order-management/product/client"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func main() {
	cfg := config.LoadConfig()
	port := cfg.GatewayPort // 8082

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("[API Gateway] Database init failed: %v", err)
	}

	prodClient := productClient.NewRemoteProductService(cfg.ProductServiceURL)
	oRepo := orderRepo.NewRepository(db)
	oSvc := orderService.NewService(oRepo, prodClient)

	// GraphQL setup
	gqlSchema, err := gatewayGraphQL.BuildGraphQLSchema(oSvc, prodClient)
	if err != nil {
		log.Fatalf("[API Gateway] GraphQL schema error: %v", err)
	}
	gqlHandler := handler.New(&handler.Config{
		Schema:   &gqlSchema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Reverse proxy targeting Order Service (Port 10004)
	targetURL, err := url.Parse(cfg.OrderServiceURL)
	if err != nil {
		log.Fatalf("[API Gateway] Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	router := gin.New()
	router.Use(
		middleware.SecurityHeaders(),
		middleware.RequestID(),
		middleware.StructuredLogger(),
		middleware.PrometheusMetrics(),
		middleware.Recovery(),
	)

	// Public Observability Endpoints
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "UP"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if err := oRepo.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	{
		// 1. Public Authentication Route (Login & JWT Generation)
		v1.POST("/auth/login", func(c *gin.Context) {
			var req LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
				return
			}

			if req.Email == "admin@example.com" && req.Password == "Admin@123" {
				token, err := middleware.GenerateToken(101, req.Email, "ADMIN")
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"token":     token,
					"tokenType": "Bearer",
					"expiresIn": 7200,
				})
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		})

		// 2. Protected Proxy Routes -> Forwarded to Order Service (Port 10004)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuthRequired())
		{
			protected.Any("/orders/*path", func(c *gin.Context) {
				proxy.ServeHTTP(c.Writer, c.Request)
			})
			protected.Any("/orders", func(c *gin.Context) {
				proxy.ServeHTTP(c.Writer, c.Request)
			})
		}
	}

	// Unified GraphQL Gateway
	router.Any("/graphql", gin.WrapH(gqlHandler))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("=====================================================")
		log.Printf("  API Gateway running on http://localhost:%s", port)
		log.Printf("=====================================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[API Gateway] Listener error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("[API Gateway] Terminated cleanly.")
}