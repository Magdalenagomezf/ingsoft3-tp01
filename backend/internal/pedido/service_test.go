package pedido

import (
	"errors"
	"net/http"
	"testing"
)

// assertBadRequest fails the test unless err is a *Error with a 400 status
// and the exact expected message.
func assertBadRequest(t *testing.T, err error, wantMessage string) {
	t.Helper()

	var pedidoErr *Error
	if !errors.As(err, &pedidoErr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if pedidoErr.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, pedidoErr.Status)
	}
	if pedidoErr.Message != wantMessage {
		t.Fatalf("expected message %q, got %q", wantMessage, pedidoErr.Message)
	}
}

// baseInput returns a PedidoInput that passes every validation, so each test
// can override only the field under test.
func baseInput() PedidoInput {
	return PedidoInput{
		ClienteNombre:   "Juan",
		ClienteContacto: "juan@mail.com",
		Items:           []PedidoItemInput{{ProductoID: 1, CantidadKg: 2}},
	}
}

// blankValues are inputs with no visible content: empty, spaces only, and a
// lone tab.
var blankValues = []struct {
	name  string
	value string
}{
	{name: "vacio", value: ""},
	{name: "solo espacios", value: "   "},
	{name: "un tabulador", value: "\t"},
}

// TestService_Create_ClienteNombreSinContenido_EsRechazado covers the
// cliente_nombre rule. A nil-db Service is safe here because validation runs
// before s.db is touched.
func TestService_Create_ClienteNombreSinContenido_EsRechazado(t *testing.T) {
	for _, tt := range blankValues {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := &Service{}
			input := baseInput()
			input.ClienteNombre = tt.value

			// Act
			_, err := s.Create(input)

			// Assert
			assertBadRequest(t, err, "cliente_nombre es requerido")
		})
	}
}

// TestService_Create_ClienteContactoSinContenido_EsRechazado covers the
// cliente_contacto rule.
func TestService_Create_ClienteContactoSinContenido_EsRechazado(t *testing.T) {
	for _, tt := range blankValues {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := &Service{}
			input := baseInput()
			input.ClienteContacto = tt.value

			// Act
			_, err := s.Create(input)

			// Assert
			assertBadRequest(t, err, "cliente_contacto es requerido")
		})
	}
}

// TestService_Create_SinItems_EsRechazado covers every way of having no
// items in one parametrized pass.
func TestService_Create_SinItems_EsRechazado(t *testing.T) {
	tests := []struct {
		name  string
		items []PedidoItemInput
	}{
		{name: "items nil", items: nil},
		{name: "items vacio", items: []PedidoItemInput{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := &Service{}
			input := baseInput()
			input.Items = tt.items

			// Act
			_, err := s.Create(input)

			// Assert
			assertBadRequest(t, err, "el pedido debe tener al menos un item")
		})
	}
}

// TestService_Create_CantidadKgInvalida_EsRechazada covers every
// non-positive cantidad_kg case, whether it's the only item or a later one,
// in one parametrized pass.
func TestService_Create_CantidadKgInvalida_EsRechazada(t *testing.T) {
	tests := []struct {
		name  string
		items []PedidoItemInput
	}{
		{
			name:  "cero en el unico item",
			items: []PedidoItemInput{{ProductoID: 1, CantidadKg: 0}},
		},
		{
			name:  "negativa",
			items: []PedidoItemInput{{ProductoID: 1, CantidadKg: -1}},
		},
		{
			name: "invalida solo en el segundo item",
			items: []PedidoItemInput{
				{ProductoID: 1, CantidadKg: 2},
				{ProductoID: 2, CantidadKg: -1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := &Service{}
			input := baseInput()
			input.Items = tt.items

			// Act
			_, err := s.Create(input)

			// Assert
			assertBadRequest(t, err, "cantidad_kg debe ser mayor a 0 en todos los items")
		})
	}
}
