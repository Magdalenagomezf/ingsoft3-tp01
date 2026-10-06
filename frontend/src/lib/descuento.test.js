import { describe, it, expect } from 'vitest';
import { descuentoMayorista } from './descuento';

describe('descuentoMayorista', () => {
  it.each([
    [1, 0],
    [10, 0],
    [10.01, 0.05],
    [25, 0.05],
    [25.01, 0.1],
    [50, 0.1],
    [50.01, 0.15],
    [200, 0.15],
  ])('con %s kg aplica un descuento de %s', (cantidadKg, esperado) => {
    // Act
    const descuento = descuentoMayorista(cantidadKg);

    // Assert
    expect(descuento).toBe(esperado);
  });

  it.each([
    ['cero', 0],
    ['negativa', -5],
    ['no numérica', NaN],
  ])('no aplica descuento cuando la cantidad es %s', (_caso, cantidadKg) => {
    // Act
    const descuento = descuentoMayorista(cantidadKg);

    // Assert
    expect(descuento).toBe(0);
  });
});
