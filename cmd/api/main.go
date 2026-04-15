package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"maincore_go/config"
	"maincore_go/middlewares"
	"maincore_go/routes"
	"maincore_go/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Configuration
	config.InitConfig()

	// Initialize Storage & Database
	config.InitDatabase()
	config.InitRedis()

	// Initialize S3 & Queue
	services.InitS3()
	services.InitQueue()

	// Start Background Worker
	go services.StartWorker()

	// Setup Gin Engine
	if config.AppConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(middlewares.CorsMiddleware())

	// Setup Socket.IO Server
	socketServer := services.InitSocketServer()
	defer socketServer.Close(nil)

	// Handle socket.io route specifically
	// zishang520/socket.io uses ServeHandler to return a http.Handler
	socketHandler := socketServer.ServeHandler(nil)
	r.GET("/socket.io/*any", gin.WrapH(socketHandler))
	r.POST("/socket.io/*any", gin.WrapH(socketHandler))


	// Health check route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to " + config.AppConfig.AppName + " (Go Edition)",
			"status":  true,
			"data": gin.H{
				"version": config.AppConfig.AppVersion,
			},
		})
	})

	// Setup Api Router Group
	api := r.Group("/api")
	routes.AuthRoutes(api)
	routes.ResetPasswordRoutes(api)
	routes.MasterRoutes(api)

	// Find available port
	originalPort := config.AppConfig.Port
	port := findAvailablePort(originalPort)
	if port != originalPort {
		log.Printf("Port %s was not available, using port %s instead", originalPort, port)
	}

	// Start server gracefully
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop backgrounds gracefully
	services.QueueServer.Stop()
	services.QueueServer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func findAvailablePort(portStr string) string {
	port, _ := strconv.Atoi(portStr)
	for {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return strconv.Itoa(port)
		}
		port++
	}
}
