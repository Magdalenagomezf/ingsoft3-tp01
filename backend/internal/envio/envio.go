// Package envio quotes the shipping cost of a wholesale order by zone and weight.
package envio

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrZonaInvalida     = errors.New("zona de envío inválida")
	ErrCantidadInvalida = errors.New("la cantidad debe ser mayor a 0")
)

// costoMinimo is the floor for any paid shipment, in ARS.
const costoMinimo = 4000

// envioGratisCABAKg is the weight from which CABA orders ship for free.
const envioGratisCABAKg = 100

type tarifa struct {
	base        float64
	porKg       float64
	diasHabiles int
}

var tarifas = map[string]tarifa{
	"caba":      {base: 3000, porKg: 150, diasHabiles: 1},
	"gba":       {base: 4500, porKg: 200, diasHabiles: 2},
	"interior":  {base: 8000, porKg: 350, diasHabiles: 4},
	"patagonia": {base: 12000, porKg: 500, diasHabiles: 6},
}

// Cotizacion is the shipping quote breakdown shown to the customer.
type Cotizacion struct {
	Zona        string
	CantidadKg  float64
	Base        float64
	PorPeso     float64
	Total       float64
	DiasHabiles int
}

// Cotizar returns the shipping quote for cantidadKg delivered to zona.
// Heavier orders get a lower per-kg rate, every paid shipment costs at least
// costoMinimo, and CABA orders of envioGratisCABAKg or more ship for free.
func Cotizar(zona string, cantidadKg float64) (Cotizacion, error) {
	if math.IsNaN(cantidadKg) || cantidadKg <= 0 {
		return Cotizacion{}, ErrCantidadInvalida
	}

	z := strings.ToLower(strings.TrimSpace(zona))
	t, ok := tarifas[z]
	if !ok {
		return Cotizacion{}, fmt.Errorf("%w: %q", ErrZonaInvalida, zona)
	}

	c := Cotizacion{Zona: z, CantidadKg: cantidadKg, DiasHabiles: t.diasHabiles}
	if cantidadKg > 100 {
		// Large orders go on a dedicated truck, which takes an extra day to book.
		c.DiasHabiles++
	}

	if z == "caba" && cantidadKg >= envioGratisCABAKg {
		return c, nil
	}

	c.Base = t.base
	c.PorPeso = redondear(t.porKg * factorPeso(cantidadKg) * cantidadKg)
	c.Total = redondear(c.Base + c.PorPeso)
	if c.Total < costoMinimo {
		c.Total = costoMinimo
	}
	return c, nil
}

// factorPeso discounts the per-kg rate by weight tier.
func factorPeso(kg float64) float64 {
	switch {
	case kg > 100:
		return 0.6
	case kg > 50:
		return 0.75
	case kg > 20:
		return 0.9
	default:
		return 1
	}
}

func redondear(v float64) float64 {
	return math.Round(v*100) / 100
}
