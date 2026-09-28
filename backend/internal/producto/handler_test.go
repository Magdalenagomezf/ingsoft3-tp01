package producto

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// decodeErrorBody decodes a {"error": "..."} response body and returns the
// message.
func decodeErrorBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error decodificando body: %v, body: %s", err, rec.Body.String())
	}
	return body["error"]
}

// newRequestWithID builds a request carrying the given path value under
// "id", as the mux would after matching "/api/productos/{id}".
func newRequestWithID(method, id string) *http.Request {
	req := httptest.NewRequest(method, "/api/productos/"+id, nil)
	req.SetPathValue("id", id)
	return req
}

// TestHandler_IDNoNumerico_Es400 covers every handler that reads {id} from
// the path, for an id that fails strconv.Atoi, in one parametrized pass.
func TestHandler_IDNoNumerico_Es400(t *testing.T) {
	tests := []struct {
		name   string
		invoke func(h *Handler, w http.ResponseWriter, r *http.Request)
	}{
		{name: "Get", invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Get(w, r) }},
		{name: "Update", invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) }},
		{name: "Delete", invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Delete(w, r) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := NewHandler(NewService(&mockRepo{}))
			req := newRequestWithID(http.MethodGet, "abc")
			rec := httptest.NewRecorder()

			// Act
			tt.invoke(h, rec, req)

			// Assert
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
			if got := decodeErrorBody(t, rec); got != "id invalido" {
				t.Fatalf("expected error %q, got %q", "id invalido", got)
			}
		})
	}
}

// TestHandler_Get_NoEncontrado_Es404 verifies Get maps sql.ErrNoRows from the
// service into a 404 with the exact message.
func TestHandler_Get_NoEncontrado_Es404(t *testing.T) {
	// Arrange
	h := NewHandler(NewService(&mockRepo{getErr: sql.ErrNoRows}))
	req := newRequestWithID(http.MethodGet, "1")
	rec := httptest.NewRecorder()

	// Act
	h.Get(rec, req)

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "producto no encontrado" {
		t.Fatalf("expected error %q, got %q", "producto no encontrado", got)
	}
}

// TestHandler_RepoFalla_Es500 covers every handler that returns a generic
// 500 when the repository fails with a non-not-found error, in one
// parametrized pass.
func TestHandler_RepoFalla_Es500(t *testing.T) {
	repoErr := errors.New("db caida")

	tests := []struct {
		name        string
		method      string
		wantMessage string
		build       func() (*Handler, *http.Request)
	}{
		{
			name:        "List",
			wantMessage: "error consultando productos",
			build: func() (*Handler, *http.Request) {
				h := NewHandler(NewService(&mockRepo{listErr: repoErr}))
				return h, httptest.NewRequest(http.MethodGet, "/api/productos", nil)
			},
		},
		{
			name:        "Get",
			wantMessage: "error consultando producto",
			build: func() (*Handler, *http.Request) {
				h := NewHandler(NewService(&mockRepo{getErr: repoErr}))
				return h, newRequestWithID(http.MethodGet, "1")
			},
		},
		{
			name:        "Create",
			wantMessage: "error creando producto",
			build: func() (*Handler, *http.Request) {
				h := NewHandler(NewService(&mockRepo{createErr: repoErr}))
				body := bytes.NewBufferString(`{"nombre":"Nuez","precio_por_kg":10,"stock_kg":5}`)
				return h, httptest.NewRequest(http.MethodPost, "/api/productos", body)
			},
		},
		{
			name:        "Update",
			wantMessage: "error actualizando producto",
			build: func() (*Handler, *http.Request) {
				h := NewHandler(NewService(&mockRepo{updateErr: repoErr}))
				body := bytes.NewBufferString(`{"nombre":"Nuez","precio_por_kg":10,"stock_kg":5}`)
				return h, newRequestWithIDAndBody(http.MethodPut, "1", body)
			},
		},
		{
			name:        "Delete",
			wantMessage: "error eliminando producto",
			build: func() (*Handler, *http.Request) {
				h := NewHandler(NewService(&mockRepo{deleteErr: repoErr}))
				return h, newRequestWithID(http.MethodDelete, "1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h, req := tt.build()
			rec := httptest.NewRecorder()

			// Act
			switch tt.name {
			case "List":
				h.List(rec, req)
			case "Get":
				h.Get(rec, req)
			case "Create":
				h.Create(rec, req)
			case "Update":
				h.Update(rec, req)
			case "Delete":
				h.Delete(rec, req)
			}

			// Assert
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
			}
			if got := decodeErrorBody(t, rec); got != tt.wantMessage {
				t.Fatalf("expected error %q, got %q", tt.wantMessage, got)
			}
		})
	}
}

// newRequestWithIDAndBody builds a request carrying both the "id" path value
// and a JSON body, as Update needs.
func newRequestWithIDAndBody(method, id string, body *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(method, "/api/productos/"+id, body)
	req.SetPathValue("id", id)
	return req
}

// TestHandler_CreateYUpdate_ValidacionInvalida_Es400ConMensaje covers
// Create and Update returning the service's exact validation message as a
// 400, in one parametrized pass.
func TestHandler_CreateYUpdate_ValidacionInvalida_Es400ConMensaje(t *testing.T) {
	invalidBody := `{"nombre":"","precio_por_kg":10,"stock_kg":5}`

	tests := []struct {
		name   string
		invoke func(h *Handler, w http.ResponseWriter, r *http.Request)
		req    func() *http.Request
	}{
		{
			name:   "Create",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Create(w, r) },
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/api/productos", bytes.NewBufferString(invalidBody))
			},
		},
		{
			name:   "Update",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) },
			req: func() *http.Request {
				return newRequestWithIDAndBody(http.MethodPut, "1", bytes.NewBufferString(invalidBody))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := NewHandler(NewService(&mockRepo{}))
			rec := httptest.NewRecorder()

			// Act
			tt.invoke(h, rec, tt.req())

			// Assert
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
			if got := decodeErrorBody(t, rec); got != ErrNombreRequerido.Error() {
				t.Fatalf("expected error %q, got %q", ErrNombreRequerido.Error(), got)
			}
		})
	}
}

// TestHandler_UpdateYDelete_ErrNotFound_Es404 covers Update and Delete
// mapping the service's ErrNotFound into a 404 with the exact message, in
// one parametrized pass.
func TestHandler_UpdateYDelete_ErrNotFound_Es404(t *testing.T) {
	validBody := `{"nombre":"Nuez","precio_por_kg":10,"stock_kg":5}`

	tests := []struct {
		name   string
		invoke func(h *Handler, w http.ResponseWriter, r *http.Request)
		req    func() *http.Request
	}{
		{
			name:   "Update",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) },
			req: func() *http.Request {
				return newRequestWithIDAndBody(http.MethodPut, "99", bytes.NewBufferString(validBody))
			},
		},
		{
			name:   "Delete",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Delete(w, r) },
			req: func() *http.Request {
				return newRequestWithID(http.MethodDelete, "99")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := NewHandler(NewService(&mockRepo{updateOk: false, deleteOk: false}))
			rec := httptest.NewRecorder()

			// Act
			tt.invoke(h, rec, tt.req())

			// Assert
			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
			}
			if got := decodeErrorBody(t, rec); got != "producto no encontrado" {
				t.Fatalf("expected error %q, got %q", "producto no encontrado", got)
			}
		})
	}
}

// TestHandler_CreateYUpdate_BodyInvalido_Es400 covers Create and Update
// rejecting malformed JSON with a 400, in one parametrized pass.
func TestHandler_CreateYUpdate_BodyInvalido_Es400(t *testing.T) {
	tests := []struct {
		name   string
		invoke func(h *Handler, w http.ResponseWriter, r *http.Request)
		req    func() *http.Request
	}{
		{
			name:   "Create",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Create(w, r) },
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/api/productos", bytes.NewBufferString(`{invalido`))
			},
		},
		{
			name:   "Update",
			invoke: func(h *Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) },
			req: func() *http.Request {
				return newRequestWithIDAndBody(http.MethodPut, "1", bytes.NewBufferString(`{invalido`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := NewHandler(NewService(&mockRepo{}))
			rec := httptest.NewRecorder()

			// Act
			tt.invoke(h, rec, tt.req())

			// Assert
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
			if got := decodeErrorBody(t, rec); got != "body invalido" {
				t.Fatalf("expected error %q, got %q", "body invalido", got)
			}
		})
	}
}

// TestHandler_CaminosFelices verifies the success status and body for List,
// Get, Create and Delete.
func TestHandler_CaminosFelices(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		// Arrange
		out := Producto{ID: 1, Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}
		h := NewHandler(NewService(&mockRepo{createOut: out}))
		body := bytes.NewBufferString(`{"nombre":"Nuez","precio_por_kg":10,"stock_kg":5}`)
		req := httptest.NewRequest(http.MethodPost, "/api/productos", body)
		rec := httptest.NewRecorder()

		// Act
		h.Create(rec, req)

		// Assert
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
		var got Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("error decodificando body: %v", err)
		}
		if got != out {
			t.Fatalf("expected body %+v, got %+v", out, got)
		}
	})

	t.Run("Get", func(t *testing.T) {
		// Arrange
		out := Producto{ID: 1, Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}
		h := NewHandler(NewService(&mockRepo{getOut: out}))
		req := newRequestWithID(http.MethodGet, "1")
		rec := httptest.NewRecorder()

		// Act
		h.Get(rec, req)

		// Assert
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var got Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("error decodificando body: %v", err)
		}
		if got != out {
			t.Fatalf("expected body %+v, got %+v", out, got)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Arrange
		h := NewHandler(NewService(&mockRepo{deleteOk: true}))
		req := newRequestWithID(http.MethodDelete, "1")
		rec := httptest.NewRecorder()

		// Act
		h.Delete(rec, req)

		// Assert
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		// Arrange
		out := []Producto{{ID: 1, Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}}
		h := NewHandler(NewService(&mockRepo{listOut: out}))
		req := httptest.NewRequest(http.MethodGet, "/api/productos", nil)
		rec := httptest.NewRecorder()

		// Act
		h.List(rec, req)

		// Assert
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var got []Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("error decodificando body: %v", err)
		}
		if len(got) != 1 || got[0] != out[0] {
			t.Fatalf("expected body %+v, got %+v", out, got)
		}
	})
}
