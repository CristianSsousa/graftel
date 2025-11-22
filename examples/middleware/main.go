package main

import (
	"context"
	"log"

	"github.com/CristianSsousa/graftel/v2"
	"github.com/gin-gonic/gin"
)

func main() {
	config := graftel.NewConfig("middleware-example").
		WithServiceVersion("1.0.0").
		WithOTLPEndpoint("http://localhost:4318").
		WithInsecure(true).
		WithResourceAttribute("environment", "development")

	client, err := graftel.NewClient(config)
	if err != nil {
		log.Fatalf("Falha ao criar cliente: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Falha ao inicializar: %v", err)
	}
	defer client.Shutdown(ctx)

	router := gin.Default()

	middlewareConfig := graftel.DefaultMiddlewareConfig("middleware-example")
	router.Use(graftel.GinMiddleware(client, middlewareConfig))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "João"},
				{"id": 2, "name": "Maria"},
			},
		})
	})

	router.POST("/api/users", func(c *gin.Context) {
		var user struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"id": 3, "name": user.Name})
	})

	router.GET("/api/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{"id": id, "name": "Usuário " + id})
	})

	router.GET("/api/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "Erro simulado"})
	})

	log.Println("Servidor iniciado em http://localhost:8080")
	log.Println("Endpoints disponíveis:")
	log.Println("  GET  /health")
	log.Println("  GET  /api/users")
	log.Println("  POST /api/users")
	log.Println("  GET  /api/users/:id")
	log.Println("  GET  /api/error")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}
