package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/galanjabal3/linkas/docs"
	"github.com/galanjabal3/linkas/internal/handler"
	"github.com/galanjabal3/linkas/internal/middleware"
	"github.com/galanjabal3/linkas/internal/repository"
	"github.com/galanjabal3/linkas/internal/service"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Linkas API
// @version         1.0
// @description     A simple, fast URL shortener API built with Go + Gin + PostgreSQL.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Galan Jabal
// @contact.url    https://github.com/galanjabal3
// @contact.email  galanjabal3@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/linkas?sslmode=disable")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	repo := repository.NewURLRepository(db)
	svc := service.NewURLService(repo)
	h := handler.NewURLHandler(svc, baseURL)

	limiter := middleware.NewRateLimiter(30, time.Minute)

	router := gin.Default()

	router.Use(middleware.RateLimit(limiter))

	// --- Swagger ---
	router.GET("/doc.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, docs.SwaggerInfo.ReadDoc())
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/doc.json")))

	// --- API routes ---
	api := router.Group("/api")
	{
		api.POST("/shorten", h.Shorten)
		api.GET("/stats/*slug", h.GetStats)
		api.GET("/qrcode/*slug", h.GetQRCode)
		api.DELETE("/*slug", h.Delete)
		api.GET("/health", h.Health)
	}

	// --- Catch-all for slug redirect (using NoRoute) ---
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Skip known prefixes
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/doc.json") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		// Strip leading slash
		slug := strings.TrimPrefix(path, "/")
		h.RedirectBySlug(c, slug)
	})

	log.Printf("Linkas running on port %s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	const (
		colorCyan  = "\033[36m"
		colorReset = "\033[0m"
		colorDim   = "\033[2m"
	)

	fmt.Println(colorCyan + `
    ___       __             
   / (_)___  / /______ ______
  / / / __ \/ //_/ __ ` + "`" + `/ ___/
 / / / / / / ,< / /_/ (__  ) 
/_/_/_/ /_/_/|_|\__,_/____/  
` + colorReset)

	fmt.Println(colorDim + "  ────────────────────────────" + colorReset)
	fmt.Println("  🔗  " + colorCyan + "URL Shortener API" + colorReset + colorDim + " v1.0" + colorReset)
	fmt.Println(colorDim + "  ────────────────────────────" + colorReset)
	fmt.Println()
}
