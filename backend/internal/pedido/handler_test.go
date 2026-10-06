package pedido

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"nueces-backend/internal/producto"
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

// TestHandler_Create_BodyInvalido_Es400 verifies malformed JSON is rejected
// before the service is ever reached (a nil service would panic otherwise).
func TestHandler_Create_BodyInvalido_Es400(t *testing.T) {
	// Arrange
	h := NewHandler(&Service{})
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", bytes.NewBufferString(`{invalido`))
	rec := httptest.NewRecorder()

	// Act
	h.Create(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "body invalido" {
		t.Fatalf("expected error %q, got %q", "body invalido", got)
	}
}

// TestHandler_Create_ValidacionInvalida_UsaStatusYMensajeDelServicio verifies
// that a *Error coming out of Service.Create (here, via a plain validation
// failure on a nil-db Service) is forwarded with its exact status and
// message.
func TestHandler_Create_ValidacionInvalida_UsaStatusYMensajeDelServicio(t *testing.T) {
	// Arrange
	h := NewHandler(&Service{})
	input := baseInput()
	input.ClienteNombre = ""
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("error serializando input: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	// Act
	h.Create(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "cliente_nombre es requerido" {
		t.Fatalf("expected error %q, got %q", "cliente_nombre es requerido", got)
	}
}

// TestHandler_List_RepoFalla_Es500 verifies List maps a repository failure
// into a 500 with the exact message.
func TestHandler_List_RepoFalla_Es500(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, cliente_nombre, cliente_contacto, fecha_creacion, estado FROM pedidos ORDER BY id`)).
		WillReturnError(sql.ErrConnDone)

	service := NewService(db, NewRepository(db), producto.NewRepository(db))
	h := NewHandler(service)
	req := httptest.NewRequest(http.MethodGet, "/api/pedidos", nil)
	rec := httptest.NewRecorder()

	// Act
	h.List(rec, req)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "error consultando pedidos" {
		t.Fatalf("expected error %q, got %q", "error consultando pedidos", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestHandler_Create_CaminoFeliz_Es201 verifies Create's happy path returns
// 201 with the created pedido as JSON, driving the real service and
// repositories over sqlmock end to end through the handler.
func TestHandler_Create_CaminoFeliz_Es201(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	defer db.Close()

	const productoID = 1

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT nombre, stock_kg, precio_por_kg FROM productos WHERE id = $1 FOR UPDATE`)).
		WithArgs(productoID).
		WillReturnRows(sqlmock.NewRows([]string{"nombre", "stock_kg", "precio_por_kg"}).
			AddRow("Nuez", 10.0, 8.5))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE productos SET stock_kg = stock_kg - $1 WHERE id = $2`)).
		WithArgs(2.0, productoID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO pedidos (cliente_nombre, cliente_contacto) VALUES ($1, $2)`)).
		WithArgs("Juan", "juan@mail.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cliente_nombre", "cliente_contacto", "fecha_creacion", "estado"}).
			AddRow(1, "Juan", "juan@mail.com", time.Now(), "pendiente"))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO pedido_items (pedido_id, producto_id, cantidad_kg, precio_unitario)`)).
		WithArgs(1, productoID, 2.0, 8.5).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	service := NewService(db, NewRepository(db), producto.NewRepository(db))
	h := NewHandler(service)

	payload, err := json.Marshal(baseInput())
	if err != nil {
		t.Fatalf("error serializando input: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	// Act
	h.Create(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var got Pedido
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("error decodificando body: %v", err)
	}
	if got.ID != 1 {
		t.Fatalf("expected pedido ID 1, got %d", got.ID)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}
