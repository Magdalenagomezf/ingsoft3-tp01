package pedido

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"nueces-backend/internal/producto"
)

// newCancelRequest builds a DELETE request carrying the given path value
// under "id", as the mux would after matching "/api/pedidos/{id}".
func newCancelRequest(id string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/api/pedidos/"+id, nil)
	req.SetPathValue("id", id)
	return req
}

// TestHandler_Cancel_IDNoNumerico_Es400 verifies a non-numeric id is rejected
// before the service is ever reached (a nil service would panic otherwise).
func TestHandler_Cancel_IDNoNumerico_Es400(t *testing.T) {
	// Arrange
	h := NewHandler(&Service{})
	rec := httptest.NewRecorder()

	// Act
	h.Cancel(rec, newCancelRequest("abc"))

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "id invalido" {
		t.Fatalf("expected error %q, got %q", "id invalido", got)
	}
}

// TestHandler_Cancel_PedidoInexistente_Es404 verifies a 404 *Error coming out
// of Service.Cancel is forwarded with its exact message.
func TestHandler_Cancel_PedidoInexistente_Es404(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	h := NewHandler(NewService(db, NewRepository(db), producto.NewRepository(db)))
	rec := httptest.NewRecorder()

	// Act
	h.Cancel(rec, newCancelRequest("99"))

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "pedido id 99 no existe" {
		t.Fatalf("expected error %q, got %q", "pedido id 99 no existe", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestHandler_Cancel_BeginFalla_Es500 verifies a transaction failure inside
// the service is forwarded as a 500 with the service message.
func TestHandler_Cancel_BeginFalla_Es500(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	h := NewHandler(NewService(db, NewRepository(db), producto.NewRepository(db)))
	rec := httptest.NewRecorder()

	// Act
	h.Cancel(rec, newCancelRequest("1"))

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if got := decodeErrorBody(t, rec); got != "error iniciando transaccion" {
		t.Fatalf("expected error %q, got %q", "error iniciando transaccion", got)
	}
}

// TestHandler_Cancel_CaminoFeliz_Es204 verifies the happy path answers 204
// with an empty body, driving the real service and repositories over sqlmock.
func TestHandler_Cancel_CaminoFeliz_Es204(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(itemsPedidoQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"producto_id", "cantidad_kg"}).AddRow(1, 2.0))
	mock.ExpectExec(regexp.QuoteMeta(devolverStockExec)).
		WithArgs(2.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(borrarItemsExec)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(borrarPedidoExec)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	h := NewHandler(NewService(db, NewRepository(db), producto.NewRepository(db)))
	rec := httptest.NewRecorder()

	// Act
	h.Cancel(rec, newCancelRequest("1"))

	// Assert
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusNoContent, rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}
