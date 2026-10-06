// Package envio quotes the shipping cost of a wholesale order shipped from
// Catamarca, by destination province and weight.
package envio

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrProvinciaInvalida = errors.New("provincia inválida")
	ErrSinEnvio          = errors.New("no hacemos envíos a esa provincia")
	ErrCantidadInvalida  = errors.New("la cantidad debe ser mayor a 0")
)

// costoMinimo is the floor for any paid shipment, in ARS.
const costoMinimo = 4000

// envioGratisLocalKg is the weight from which orders within Catamarca ship for free.
const envioGratisLocalKg = 50

// zona groups provinces by distance from Catamarca.
type zona string

const (
	zonaLocal       zona = "Local"
	zonaNOA         zona = "NOA"
	zonaCentro      zona = "Centro"
	zonaAMBALitoral zona = "AMBA/Litoral"
	zonaSinEnvio    zona = "Sin envío"
)

// provincias maps each normalized province name to its shipping zone.
var provincias = map[string]zona{
	"catamarca": zonaLocal,

	"tucuman":             zonaNOA,
	"santiago del estero": zonaNOA,
	"la rioja":            zonaNOA,
	"salta":               zonaNOA,
	"jujuy":               zonaNOA,

	"cordoba":  zonaCentro,
	"santa fe": zonaCentro,
	"san luis": zonaCentro,
	"san juan": zonaCentro,
	"mendoza":  zonaCentro,

	"buenos aires":                    zonaAMBALitoral,
	"caba":                            zonaAMBALitoral,
	"ciudad autonoma de buenos aires": zonaAMBALitoral,
	"entre rios":                      zonaAMBALitoral,
	"corrientes":                      zonaAMBALitoral,
	"misiones":                        zonaAMBALitoral,
	"chaco":                           zonaAMBALitoral,
	"formosa":                         zonaAMBALitoral,

	"la pampa":         zonaSinEnvio,
	"neuquen":          zonaSinEnvio,
	"rio negro":        zonaSinEnvio,
	"chubut":           zonaSinEnvio,
	"santa cruz":       zonaSinEnvio,
	"tierra del fuego": zonaSinEnvio,
}

type tarifa struct {
	base        float64
	porKg       float64
	diasHabiles int
}

var tarifas = map[zona]tarifa{
	zonaLocal:       {base: 2500, porKg: 100, diasHabiles: 1},
	zonaNOA:         {base: 5000, porKg: 250, diasHabiles: 2},
	zonaCentro:      {base: 8000, porKg: 350, diasHabiles: 3},
	zonaAMBALitoral: {base: 11000, porKg: 450, diasHabiles: 4},
}

// Cotizacion is the shipping quote breakdown shown to the customer.
type Cotizacion struct {
	Provincia   string
	Zona        string
	CantidadKg  float64
	Base        float64
	PorPeso     float64
	Total       float64
	DiasHabiles int
}

// Cotizar returns the shipping quote for cantidadKg delivered to provincia.
// Heavier orders get a lower per-kg rate, every paid shipment costs at least
// costoMinimo, and orders within Catamarca of envioGratisLocalKg or more ship
// for free.
func Cotizar(provincia string, cantidadKg float64) (Cotizacion, error) {
	if math.IsNaN(cantidadKg) || cantidadKg <= 0 {
		return Cotizacion{}, ErrCantidadInvalida
	}

	p := normalizar(provincia)
	z, ok := provincias[p]
	if !ok {
		return Cotizacion{}, fmt.Errorf("%w: %q", ErrProvinciaInvalida, provincia)
	}
	if z == zonaSinEnvio {
		return Cotizacion{}, fmt.Errorf("%w: %q", ErrSinEnvio, provincia)
	}

	t := tarifas[z]
	c := Cotizacion{Provincia: p, Zona: string(z), CantidadKg: cantidadKg, DiasHabiles: t.diasHabiles}
	if cantidadKg > 100 {
		// Large orders go on a dedicated truck, which takes an extra day to book.
		c.DiasHabiles++
	}

	if z == zonaLocal && cantidadKg >= envioGratisLocalKg {
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

// sinTildes strips the Spanish accents that can appear in province names.
var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u")

// normalizar lowercases the province name, strips accents and collapses
// whitespace, so "  Entre   Ríos " and "entre rios" match the same key.
func normalizar(provincia string) string {
	p := sinTildes.Replace(strings.ToLower(provincia))
	return strings.Join(strings.Fields(p), " ")
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
