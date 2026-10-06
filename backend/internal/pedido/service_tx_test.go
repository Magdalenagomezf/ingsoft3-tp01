package pedido

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"nueces-backend/internal/producto"
)

// newTxTestService builds a Service backed by real repositories over a
// sqlmock database, so the transactional Create path (Begin, locks, updates,
// inserts, Commit/Rollback) runs against actual SQL rather than a hand
// written fake.
func newTxTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("error creando sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewRepository(db)
	productoRepo := producto.NewRepository(db)
	service := NewService(db, repo, productoRepo)

	return service, mock
}

// TestService_Create_StockInsuficiente_EsRechazado verifies that when the
// locked producto doesn't have enough stock, Create rolls back the
// transaction and returns the exact 400 message, without ever touching
// UPDATE/INSERT.
func TestService_Create_StockInsuficiente_EsRechazado(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT nombre, stock_kg, precio_por_kg FROM productos WHERE id = $1 FOR UPDATE`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"nombre", "stock_kg", "precio_por_kg"}).
			AddRow("Nuez", 5.0, 10.0))
	mock.ExpectRollback()

	input := baseInput()
	input.Items = []PedidoItemInput{{ProductoID: 1, CantidadKg: 100}}

	// Act
	_, err := service.Create(input)

	// Assert
	wantMessage := fmt.Sprintf("stock insuficiente para: %s (pedido: %.2fkg, disponible: %.2fkg)", "Nuez", 100.0, 5.0)
	assertBadRequest(t, err, wantMessage)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Create_ProductoInexistente_EsRechazado verifies that when the
// lock query reports no rows for a producto id, Create rolls back and
// returns the exact 400 message identifying that id.
func TestService_Create_ProductoInexistente_EsRechazado(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT nombre, stock_kg, precio_por_kg FROM productos WHERE id = $1 FOR UPDATE`)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	input := baseInput()
	input.Items = []PedidoItemInput{{ProductoID: 99, CantidadKg: 1}}

	// Act
	_, err := service.Create(input)

	// Assert
	assertBadRequest(t, err, "producto id 99 no existe")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Create_CaminoFeliz_ConfirmaLaTransaccion verifies the full
// happy path: the lock is read, stock is decremented by the summed cantidad
// of two items of the same producto, the pedido and its items are inserted,
// and the transaction is committed. It also asserts the returned Pedido
// carries the locked row's nombre and precio.
func TestService_Create_CaminoFeliz_ConfirmaLaTransaccion(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	const productoID = 1
	const stockDisponible = 10.0
	const precioPorKg = 8.5
	const primerItemKg = 1.5
	const segundoItemKg = 0.5
	const cantidadSumada = primerItemKg + segundoItemKg

	fechaCreacion := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT nombre, stock_kg, precio_por_kg FROM productos WHERE id = $1 FOR UPDATE`)).
		WithArgs(productoID).
		WillReturnRows(sqlmock.NewRows([]string{"nombre", "stock_kg", "precio_por_kg"}).
			AddRow("Nuez", stockDisponible, precioPorKg))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE productos SET stock_kg = stock_kg - $1 WHERE id = $2`)).
		WithArgs(cantidadSumada, productoID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO pedidos (cliente_nombre, cliente_contacto) VALUES ($1, $2)`)).
		WithArgs("Juan", "juan@mail.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cliente_nombre", "cliente_contacto", "fecha_creacion", "estado"}).
			AddRow(1, "Juan", "juan@mail.com", fechaCreacion, "pendiente"))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO pedido_items (pedido_id, producto_id, cantidad_kg, precio_unitario)`)).
		WithArgs(1, productoID, primerItemKg, precioPorKg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO pedido_items (pedido_id, producto_id, cantidad_kg, precio_unitario)`)).
		WithArgs(1, productoID, segundoItemKg, precioPorKg).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	input := baseInput()
	input.Items = []PedidoItemInput{
		{ProductoID: productoID, CantidadKg: primerItemKg},
		{ProductoID: productoID, CantidadKg: segundoItemKg},
	}

	// Act
	result, err := service.Create(input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Fatalf("expected pedido ID 1, got %d", result.ID)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	for _, item := range result.Items {
		if item.ProductoNombre != "Nuez" {
			t.Fatalf("expected ProductoNombre %q, got %q", "Nuez", item.ProductoNombre)
		}
		if item.PrecioUnitario != precioPorKg {
			t.Fatalf("expected PrecioUnitario %v, got %v", precioPorKg, item.PrecioUnitario)
		}
	}
	if result.Items[0].CantidadKg != primerItemKg || result.Items[1].CantidadKg != segundoItemKg {
		t.Fatalf("expected item cantidades %v and %v, got %v and %v", primerItemKg, segundoItemKg, result.Items[0].CantidadKg, result.Items[1].CantidadKg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Create_BeginFalla_Retorna500 verifies that when starting the
// transaction fails, Create returns a 500 with the exact wording and never
// attempts a rollback (there's no transaction to roll back).
func TestService_Create_BeginFalla_Retorna500(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	mock.ExpectBegin().WillReturnError(errors.New("conexion caida"))

	input := baseInput()

	// Act
	_, err := service.Create(input)

	// Assert
	var pedidoErr *Error
	if !errors.As(err, &pedidoErr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if pedidoErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, pedidoErr.Status)
	}
	if pedidoErr.Message != "error iniciando transaccion" {
		t.Fatalf("expected message %q, got %q", "error iniciando transaccion", pedidoErr.Message)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}
