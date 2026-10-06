// Regla comercial del mayorista: descuento por volumen según los kg del pedido.

/**
 * Devuelve el descuento mayorista que corresponde a una cantidad en kg.
 * Escalones: hasta 10 kg → 0; más de 10 → 5%; más de 25 → 10%; más de 50 → 15%.
 * Una cantidad no positiva (o que no es un número) no tiene descuento.
 * @param {number} cantidadKg
 * @returns {number} descuento como fracción (0.05 = 5%)
 */
export function descuentoMayorista(cantidadKg) {
  if (!(cantidadKg > 0)) return 0;
  if (cantidadKg > 50) return 0.15;
  if (cantidadKg > 25) return 0.1;
  if (cantidadKg > 10) return 0.05;
  return 0;
}
