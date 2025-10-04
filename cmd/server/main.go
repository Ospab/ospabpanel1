package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"ospab-panel/internal/api"
	"ospab-panel/internal/core/server"
	"ospab-panel/internal/core/user"
	"ospab-panel/internal/hypervisor"
	"ospab-panel/internal/infra/db"
	"ospab-panel/pkg/auth"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Инициализация базы данных
	repository, err := db.NewRepository()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.Close()

	// Инициализация сервисов
	userService := user.NewService(repository.GetDB())
	serverService := server.NewService(repository.GetDB())
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"))
	// Гипервизоры
	hvFactory := hypervisor.NewHypervisorFactory()

	// Инициализация API обработчиков
	apiHandler := api.NewHandler(userService, serverService, hvFactory, jwtManager)

	// Создание роутеров
	apiRouter := apiHandler.SetupRoutes()
	webRouter := setupWebRoutes()

	// Запуск серверов
	go startAPIServer(apiRouter)
	go startWebServer(webRouter)

	log.Println("OSPAB Panel 0.1-alpha started successfully")
	log.Printf("API Server: http://localhost:%s", os.Getenv("API_PORT"))
	log.Printf("Web Server: http://localhost:%s", os.Getenv("WEB_PORT"))

	// Graceful shutdown
	waitForShutdown()
}

func startAPIServer(router *mux.Router) {
	port := getEnvOrDefault("API_PORT", "5000")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("API server failed: %v", err)
	}
}

func startWebServer(router *mux.Router) {
	port := getEnvOrDefault("WEB_PORT", "3000")
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("Port %s busy, trying next free port...", port)
		// Авто-подбор свободного
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			log.Fatalf("Failed to bind web server: %v", err)
		}
		addr := ln.Addr().String()
		_, autoPort, _ := net.SplitHostPort(addr)
		port = autoPort
	}
	server := &http.Server{
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("Web server starting on port %s", port)
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Web server failed: %v", err)
		}
	}()
}

func setupWebRoutes() *mux.Router {
	r := mux.NewRouter()
	// Определяем директорию статических файлов: приоритет dist (собранный React), иначе fallback на static
	distDir := "web/dist"
	staticDir := "web/static"
	var serveDir string
	if info, err := os.Stat(distDir); err == nil && info.IsDir() {
		serveDir = distDir
	} else {
		serveDir = staticDir
	}
	abs, _ := filepath.Abs(serveDir)
	log.Printf("Serving frontend from: %s", abs)
	fs := http.FileServer(http.Dir(serveDir))
	r.PathPrefix("/").Handler(fs)
	return r
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Здесь можно добавить graceful shutdown логику
	_ = ctx

	log.Println("Servers stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
