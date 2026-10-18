package pedido

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	lockPedidoQuery   = `SELECT id FROM pedidos WHERE id = $1 FOR UPDATE`
	itemsPedidoQuery  = `SELECT producto_id, cantidad_kg FROM pedido_items WHERE pedido_id = $1`
	devolverStockExec = `UPDATE productos SET stock_kg = stock_kg + $1 WHERE id = $2`
	borrarItemsExec   = `DELETE FROM pedido_items WHERE pedido_id = $1`
	borrarPedidoExec  = `DELETE FROM pedidos WHERE id = $1`
)

// assertCancelError fails the test unless err is a *Error with the expected
// status and exact message.
func assertCancelError(t *testing.T, err error, wantStatus int, wantMessage string) {
	t.Helper()

	var pedidoErr *Error
	if !errors.As(err, &pedidoErr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if pedidoErr.Status != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, pedidoErr.Status)
	}
	if pedidoErr.Message != wantMessage {
		t.Fatalf("expected message %q, got %q", wantMessage, pedidoErr.Message)
	}
}

// TestService_Cancel_CaminoFeliz_DevuelveStockYConfirma verifies the full
// cancellation path: the pedido is locked, its items are read, stock is
// returned summed per producto, items and pedido are deleted, and the
// transaction is committed.
func TestService_Cancel_CaminoFeliz_DevuelveStockYConfirma(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	const pedidoID = 7
	const productoID = 1

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).
		WithArgs(pedidoID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pedidoID))
	mock.ExpectQuery(regexp.QuoteMeta(itemsPedidoQuery)).
		WithArgs(pedidoID).
		WillReturnRows(sqlmock.NewRows([]string{"producto_id", "cantidad_kg"}).
			AddRow(productoID, 1.5).
			AddRow(productoID, 0.5))
	mock.ExpectExec(regexp.QuoteMeta(devolverStockExec)).
		WithArgs(2.0, productoID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(borrarItemsExec)).
		WithArgs(pedidoID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(borrarPedidoExec)).
		WithArgs(pedidoID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err := service.Cancel(pedidoID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Cancel_PedidoInexistente_Es404 verifies that when the pedido
// lock reports no rows, Cancel rolls back and returns a 404.
func TestService_Cancel_PedidoInexistente_Es404(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	// Act
	err := service.Cancel(99)

	// Assert
	assertCancelError(t, err, http.StatusNotFound, "pedido id 99 no existe")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Cancel_BeginFalla_Retorna500 verifies that a failure starting
// the transaction yields a 500 without any rollback.
func TestService_Cancel_BeginFalla_Retorna500(t *testing.T) {
	// Arrange
	service, mock := newTxTestService(t)

	mock.ExpectBegin().WillReturnError(errors.New("conexion caida"))

	// Act
	err := service.Cancel(1)

	// Assert
	assertCancelError(t, err, http.StatusInternalServerError, "error iniciando transaccion")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
	}
}

// TestService_Cancel_FallaLaBase_HaceRollback covers every database failure
// after Begin: each one must roll back the transaction (so stock and pedido
// stay untouched) and surface a 500 with the exact wording.
func TestService_Cancel_FallaLaBase_HaceRollback(t *testing.T) {
	const pedidoID = 1
	dbErr := errors.New("fallo de base")

	lockOK := func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).WithArgs(pedidoID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pedidoID))
	}
	itemsOK := func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(regexp.QuoteMeta(itemsPedidoQuery)).WithArgs(pedidoID).
			WillReturnRows(sqlmock.NewRows([]string{"producto_id", "cantidad_kg"}).AddRow(1, 2.0))
	}
	stockOK := func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta(devolverStockExec)).WithArgs(2.0, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	borrarItemsOK := func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta(borrarItemsExec)).WithArgs(pedidoID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	borrarPedidoOK := func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(regexp.QuoteMeta(borrarPedidoExec)).WithArgs(pedidoID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	tests := []struct {
		name        string
		setup       func(mock sqlmock.Sqlmock)
		wantMessage string
	}{
		{
			name: "falla el lock del pedido",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(lockPedidoQuery)).WithArgs(pedidoID).WillReturnError(dbErr)
			},
			wantMessage: "error consultando pedido",
		},
		{
			name: "falla la lectura de items",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				mock.ExpectQuery(regexp.QuoteMeta(itemsPedidoQuery)).WithArgs(pedidoID).WillReturnError(dbErr)
			},
			wantMessage: "error consultando items del pedido",
		},
		{
			name: "falla el escaneo de una fila de items",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				mock.ExpectQuery(regexp.QuoteMeta(itemsPedidoQuery)).WithArgs(pedidoID).
					WillReturnRows(sqlmock.NewRows([]string{"producto_id", "cantidad_kg"}).AddRow("no-es-un-numero", 1.0))
			},
			wantMessage: "error consultando items del pedido",
		},
		{
			name: "falla la devolucion de stock",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				itemsOK(mock)
				mock.ExpectExec(regexp.QuoteMeta(devolverStockExec)).WithArgs(2.0, 1).WillReturnError(dbErr)
			},
			wantMessage: "error devolviendo stock",
		},
		{
			name: "falla el borrado de items",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				itemsOK(mock)
				stockOK(mock)
				mock.ExpectExec(regexp.QuoteMeta(borrarItemsExec)).WithArgs(pedidoID).WillReturnError(dbErr)
			},
			wantMessage: "error eliminando items del pedido",
		},
		{
			name: "falla el borrado del pedido",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				itemsOK(mock)
				stockOK(mock)
				borrarItemsOK(mock)
				mock.ExpectExec(regexp.QuoteMeta(borrarPedidoExec)).WithArgs(pedidoID).WillReturnError(dbErr)
			},
			wantMessage: "error eliminando pedido",
		},
		{
			name: "falla el commit",
			setup: func(mock sqlmock.Sqlmock) {
				lockOK(mock)
				itemsOK(mock)
				stockOK(mock)
				borrarItemsOK(mock)
				borrarPedidoOK(mock)
				mock.ExpectCommit().WillReturnError(dbErr)
			},
			wantMessage: "error confirmando cancelacion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, mock := newTxTestService(t)
			mock.ExpectBegin()
			tt.setup(mock)
			// After a failed Commit the transaction is already finished, so
			// the deferred Rollback never reaches the driver.
			if tt.wantMessage != "error confirmando cancelacion" {
				mock.ExpectRollback()
			}

			// Act
			err := service.Cancel(pedidoID)

			// Assert
			assertCancelError(t, err, http.StatusInternalServerError, tt.wantMessage)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectativas de sqlmock no cumplidas: %v", err)
			}
		})
	}
}
