package graftel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func createTestClientForMiddleware() Client {
	config := NewConfig("test-service").WithInsecure(true)
	client, err := NewClient(config)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		panic(err)
	}
	return client
}

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig("test-service")
	if config.ServiceName != "test-service" {
		t.Fatalf("ServiceName incorreto: esperado 'test-service', obtido '%s'", config.ServiceName)
	}
	if len(config.SkipPaths) == 0 {
		t.Fatal("SkipPaths deve ter valores padrão")
	}
}

func TestShouldSkip(t *testing.T) {
	skipPaths := []string{"/health", "/metrics"}
	if !shouldSkip("/health", skipPaths) {
		t.Fatal("shouldSkip deve retornar true para /health")
	}
	if !shouldSkip("/metrics", skipPaths) {
		t.Fatal("shouldSkip deve retornar true para /metrics")
	}
	if shouldSkip("/api/users", skipPaths) {
		t.Fatal("shouldSkip deve retornar false para /api/users")
	}
}

func TestHTTPMiddleware(t *testing.T) {
	client := createTestClientForMiddleware()
	defer func() {
		if err := client.Shutdown(context.Background()); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()

	config := DefaultMiddlewareConfig("test-service")
	middleware := HTTPMiddleware(client, config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Logf("Erro ao escrever resposta: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code incorreto: esperado %d, obtido %d", http.StatusOK, w.Code)
	}
}

func TestHTTPMiddleware_SkipPath(t *testing.T) {
	client := createTestClientForMiddleware()
	defer func() {
		if err := client.Shutdown(context.Background()); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()

	config := DefaultMiddlewareConfig("test-service")
	middleware := HTTPMiddleware(client, config)

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("Handler deve ser chamado mesmo para paths skipados")
	}
}

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := createTestClientForMiddleware()
	defer func() {
		if err := client.Shutdown(context.Background()); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()

	config := DefaultMiddlewareConfig("test-service")
	middleware := GinMiddleware(client, config)

	router := gin.New()
	router.Use(middleware)
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code incorreto: esperado %d, obtido %d", http.StatusOK, w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	client := createTestClientForMiddleware()
	defer func() {
		if err := client.Shutdown(context.Background()); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()

	config := DefaultMiddlewareConfig("test-service")
	middleware := EchoMiddleware(client, config)

	e := echo.New()
	e.Use(middleware)
	e.GET("/api/test", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "OK"})
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code incorreto: esperado %d, obtido %d", http.StatusOK, w.Code)
	}
}

func TestChiMiddleware(t *testing.T) {
	client := createTestClientForMiddleware()
	defer func() {
		if err := client.Shutdown(context.Background()); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()

	config := DefaultMiddlewareConfig("test-service")
	middleware := ChiMiddleware(client, config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Logf("Erro ao escrever resposta: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code incorreto: esperado %d, obtido %d", http.StatusOK, w.Code)
	}
}
