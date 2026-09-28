package producto

import (
	"errors"
	"strings"
)

var (
	ErrNombreRequerido = errors.New("nombre es requerido")
	ErrPrecioInvalido  = errors.New("precio_por_kg debe ser mayor a 0")
	ErrStockNegativo   = errors.New("stock_kg no puede ser negativo")
	ErrNotFound        = errors.New("producto no encontrado")
)

// Repo is the subset of Repository's behavior Service depends on. It
// excludes LockForUpdate and DecrementStock, which are used only by the
// pedido package directly against the concrete *Repository.
type Repo interface {
	List() ([]Producto, error)
	Get(id int) (Producto, error)
	Create(p Producto) (Producto, error)
	Update(id int, p Producto) (bool, error)
	Delete(id int) (bool, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func validate(p Producto) error {
	if strings.TrimSpace(p.Nombre) == "" {
		return ErrNombreRequerido
	}
	if p.PrecioPorKg <= 0 {
		return ErrPrecioInvalido
	}
	if p.StockKg < 0 {
		return ErrStockNegativo
	}
	return nil
}

func (s *Service) List() ([]Producto, error) {
	return s.repo.List()
}

func (s *Service) Get(id int) (Producto, error) {
	return s.repo.Get(id)
}

func (s *Service) Create(p Producto) (Producto, error) {
	if err := validate(p); err != nil {
		return Producto{}, err
	}
	return s.repo.Create(p)
}

func (s *Service) Update(id int, p Producto) (Producto, error) {
	if err := validate(p); err != nil {
		return Producto{}, err
	}
	ok, err := s.repo.Update(id, p)
	if err != nil {
		return Producto{}, err
	}
	if !ok {
		return Producto{}, ErrNotFound
	}
	p.ID = id
	return p, nil
}

func (s *Service) Delete(id int) error {
	ok, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}
