package producto

import (
	"errors"
	"testing"
)

// mockRepo is a hand-written mock implementing Repo, used to verify Service
// behavior without touching a real database.
type mockRepo struct {
	listOut []Producto
	listErr error

	getCalls int
	getArg   int
	getOut   Producto
	getErr   error

	createCalls int
	createArg   Producto
	createErr   error
	createOut   Producto

	updateCalls int
	updateArgID int
	updateOk    bool
	updateErr   error

	deleteCalls int
	deleteArg   int
	deleteOk    bool
	deleteErr   error
}

func (m *mockRepo) List() ([]Producto, error) {
	return m.listOut, m.listErr
}

func (m *mockRepo) Get(id int) (Producto, error) {
	m.getCalls++
	m.getArg = id
	return m.getOut, m.getErr
}

func (m *mockRepo) Create(p Producto) (Producto, error) {
	m.createCalls++
	m.createArg = p
	if m.createErr != nil {
		return Producto{}, m.createErr
	}
	return m.createOut, nil
}

func (m *mockRepo) Update(id int, p Producto) (bool, error) {
	m.updateCalls++
	m.updateArgID = id
	return m.updateOk, m.updateErr
}

func (m *mockRepo) Delete(id int) (bool, error) {
	m.deleteCalls++
	m.deleteArg = id
	return m.deleteOk, m.deleteErr
}

// TestValidate_NombreSinContenido_EsRechazado covers every nombre with no
// visible content in one parametrized pass.
func TestValidate_NombreSinContenido_EsRechazado(t *testing.T) {
	tests := []struct {
		name   string
		nombre string
	}{
		{name: "vacio", nombre: ""},
		{name: "solo espacios", nombre: "   "},
		{name: "un tabulador", nombre: "\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := Producto{Nombre: tt.nombre, PrecioPorKg: 10, StockKg: 5}

			// Act
			err := validate(p)

			// Assert
			if !errors.Is(err, ErrNombreRequerido) {
				t.Fatalf("expected ErrNombreRequerido, got %v", err)
			}
		})
	}
}

// TestValidate_PrecioInvalido_EsRechazado covers every non-positive precio
// value in one parametrized pass.
func TestValidate_PrecioInvalido_EsRechazado(t *testing.T) {
	tests := []struct {
		name   string
		precio float64
	}{
		{name: "cero", precio: 0},
		{name: "negativo", precio: -1},
		{name: "centavo negativo", precio: -0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := Producto{Nombre: "Nuez", PrecioPorKg: tt.precio, StockKg: 5}

			// Act
			err := validate(p)

			// Assert
			if !errors.Is(err, ErrPrecioInvalido) {
				t.Fatalf("expected ErrPrecioInvalido, got %v", err)
			}
		})
	}
}

// TestValidate_StockNegativo_EsRechazado covers every negative stock value
// in one parametrized pass.
func TestValidate_StockNegativo_EsRechazado(t *testing.T) {
	tests := []struct {
		name  string
		stock float64
	}{
		{name: "centavo negativo", stock: -0.01},
		{name: "uno negativo", stock: -1},
		{name: "cien negativo", stock: -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := Producto{Nombre: "Nuez", PrecioPorKg: 10, StockKg: tt.stock}

			// Act
			err := validate(p)

			// Assert
			if !errors.Is(err, ErrStockNegativo) {
				t.Fatalf("expected ErrStockNegativo, got %v", err)
			}
		})
	}
}

// TestValidate_ValoresEnElBorde_SonAceptados covers the boundary values that
// must pass validation, since they sit right at the edge of a rejected range.
func TestValidate_ValoresEnElBorde_SonAceptados(t *testing.T) {
	tests := []struct {
		name string
		p    Producto
	}{
		{name: "stock cero", p: Producto{Nombre: "Nuez", PrecioPorKg: 10, StockKg: 0}},
		{name: "precio minimo positivo", p: Producto{Nombre: "Nuez", PrecioPorKg: 0.01, StockKg: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := tt.p

			// Act
			err := validate(p)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestService_Create_ProductoInvalido_NoLlamaAlRepo verifies validation
// short-circuits Create before the repository is ever touched.
func TestService_Create_ProductoInvalido_NoLlamaAlRepo(t *testing.T) {
	// Arrange
	repo := &mockRepo{}
	s := NewService(repo)
	invalid := Producto{Nombre: "", PrecioPorKg: 10, StockKg: 5}

	// Act
	_, err := s.Create(invalid)

	// Assert
	if !errors.Is(err, ErrNombreRequerido) {
		t.Fatalf("expected ErrNombreRequerido, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create to be called 0 times, got %d", repo.createCalls)
	}
}

// TestService_Create_ProductoValido_LlamaAlRepoUnaVez verifies Create
// delegates to the repository exactly once with the given producto and
// returns its result.
func TestService_Create_ProductoValido_LlamaAlRepoUnaVez(t *testing.T) {
	// Arrange
	valid := Producto{Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}
	repo := &mockRepo{createOut: Producto{ID: 1, Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}}
	s := NewService(repo)

	// Act
	result, err := s.Create(valid)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create to be called 1 time, got %d", repo.createCalls)
	}
	if repo.createArg != valid {
		t.Fatalf("expected repo.Create to receive %+v, got %+v", valid, repo.createArg)
	}
	if result.ID != 1 {
		t.Fatalf("expected returned producto to have ID 1, got %d", result.ID)
	}
}

// TestService_Update_ProductoInvalido_NoLlamaAlRepo verifies validation
// short-circuits Update before the repository is ever touched.
func TestService_Update_ProductoInvalido_NoLlamaAlRepo(t *testing.T) {
	// Arrange
	repo := &mockRepo{}
	s := NewService(repo)
	invalid := Producto{Nombre: "Nuez", PrecioPorKg: 0, StockKg: 5}

	// Act
	_, err := s.Update(1, invalid)

	// Assert
	if !errors.Is(err, ErrPrecioInvalido) {
		t.Fatalf("expected ErrPrecioInvalido, got %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected repo.Update to be called 0 times, got %d", repo.updateCalls)
	}
}

// TestService_Update_RepoNoEncuentra_RetornaErrNotFound verifies that when
// the repository reports no rows were affected, Service translates it into
// ErrNotFound.
func TestService_Update_RepoNoEncuentra_RetornaErrNotFound(t *testing.T) {
	// Arrange
	valid := Producto{Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}
	repo := &mockRepo{updateOk: false, updateErr: nil}
	s := NewService(repo)

	// Act
	_, err := s.Update(99, valid)

	// Assert
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected repo.Update to be called 1 time, got %d", repo.updateCalls)
	}
}

// TestService_Update_Exitoso_DevuelveElProductoConElID verifies that on a
// successful repo update, Service stamps the given id onto the returned
// producto.
func TestService_Update_Exitoso_DevuelveElProductoConElID(t *testing.T) {
	// Arrange
	valid := Producto{Nombre: "Nuez", PrecioPorKg: 10, StockKg: 5}
	repo := &mockRepo{updateOk: true, updateErr: nil}
	s := NewService(repo)

	// Act
	result, err := s.Update(7, valid)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 7 {
		t.Fatalf("expected result.ID 7, got %d", result.ID)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected repo.Update to be called 1 time, got %d", repo.updateCalls)
	}
	if repo.updateArgID != 7 {
		t.Fatalf("expected repo.Update to receive id 7, got %d", repo.updateArgID)
	}
}

// TestService_Delete_RepoNoEncuentra_RetornaErrNotFound verifies that when
// the repository reports no rows were affected, Service translates it into
// ErrNotFound.
func TestService_Delete_RepoNoEncuentra_RetornaErrNotFound(t *testing.T) {
	// Arrange
	repo := &mockRepo{deleteOk: false, deleteErr: nil}
	s := NewService(repo)

	// Act
	err := s.Delete(99)

	// Assert
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestService_Delete_RepoElimina_RetornaNil verifies that on a successful
// repo delete, Service returns nil and forwards the given id.
func TestService_Delete_RepoElimina_RetornaNil(t *testing.T) {
	// Arrange
	repo := &mockRepo{deleteOk: true, deleteErr: nil}
	s := NewService(repo)

	// Act
	err := s.Delete(5)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected repo.Delete to be called 1 time, got %d", repo.deleteCalls)
	}
	if repo.deleteArg != 5 {
		t.Fatalf("expected repo.Delete to receive id 5, got %d", repo.deleteArg)
	}
}
