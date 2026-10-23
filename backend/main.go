package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"nueces-backend/internal/httpx"
	"nueces-backend/internal/pedido"
	"nueces-backend/internal/producto"
)

func main() {
	// Sentry: el DSN llega por variable de entorno (sólo está cargada en PROD).
	// Si no está, la librería queda apagada y la aplicación anda igual.
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		EnableTracing:    true,
		TracesSampleRate: 1.0, // medir todos los pedidos
	}); err != nil {
		log.Printf("no se pudo iniciar Sentry: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	db, err := connectDB()
	if err != nil {
		log.Fatalf("no se pudo inicializar la base de datos: %v", err)
	}
	defer db.Close()

	productoRepo := producto.NewRepository(db)
	productoService := producto.NewService(productoRepo)
	productoHandler := producto.NewHandler(productoService)

	pedidoRepo := pedido.NewRepository(db)
	pedidoService := pedido.NewService(db, pedidoRepo, productoRepo)
	pedidoHandler := pedido.NewHandler(pedidoService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)

	mux.HandleFunc("GET /api/productos", productoHandler.List)
	mux.HandleFunc("GET /api/productos/{id}", productoHandler.Get)
	mux.HandleFunc("POST /api/productos", productoHandler.Create)
	mux.HandleFunc("PUT /api/productos/{id}", productoHandler.Update)
	mux.HandleFunc("DELETE /api/productos/{id}", productoHandler.Delete)

	mux.HandleFunc("POST /api/pedidos", pedidoHandler.Create)
	mux.HandleFunc("GET /api/pedidos", pedidoHandler.List)
	mux.HandleFunc("DELETE /api/pedidos/{id}", pedidoHandler.Cancel)

	

	// Cada pedido pasa primero por Sentry: mide cuánto tarda y reporta si algo explota.
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})

	log.Println("servidor escuchando en :8080")
	if err := http.ListenAndServe(":8080", sentryHandler.Handle(mux)); err != nil {
		log.Fatalf("error en el servidor: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

