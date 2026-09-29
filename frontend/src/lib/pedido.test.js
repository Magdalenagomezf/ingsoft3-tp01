import { describe, it, expect, vi } from 'vitest';
import { validarPedido, agregarItem, totalCarrito, confirmarPedido } from './pedido';

const NOMBRE_VALIDO = 'Ana';
const CONTACTO_VALIDO = 'ana@mail.com';
const ITEM_VALIDO = { producto_id: 1, nombre: 'Nueces', precio_por_kg: 100, cantidad_kg: 2 };

function carritoValido() {
  return [{ ...ITEM_VALIDO }];
}

describe('validarPedido', () => {
  it.each([
    ['vacío', ''],
    ['sólo espacios', '   '],
    ['un tabulador', '\t'],
  ])('devuelve error cuando el nombre está %s', (_caso, nombre) => {
    // Arrange
    const carrito = carritoValido();

    // Act
    const errores = validarPedido(nombre, CONTACTO_VALIDO, carrito);

    // Assert
    expect(errores).toContain('El nombre del cliente es requerido.');
  });

  it.each([
    ['vacío', ''],
    ['sólo espacios', '   '],
    ['un tabulador', '\t'],
  ])('devuelve error cuando el contacto está %s', (_caso, contacto) => {
    // Arrange
    const carrito = carritoValido();

    // Act
    const errores = validarPedido(NOMBRE_VALIDO, contacto, carrito);

    // Assert
    expect(errores).toContain('El contacto del cliente es requerido.');
  });

  it('devuelve exactamente el error de carrito vacío cuando no hay productos', () => {
    // Arrange
    const carrito = [];

    // Act
    const errores = validarPedido(NOMBRE_VALIDO, CONTACTO_VALIDO, carrito);

    // Assert
    expect(errores).toEqual(['Agregá al menos un producto al pedido.']);
  });

  it.each([
    ['cantidad 0', [{ ...ITEM_VALIDO, cantidad_kg: 0 }]],
    ['cantidad negativa', [{ ...ITEM_VALIDO, cantidad_kg: -1 }]],
    ['cantidad undefined', [{ ...ITEM_VALIDO, cantidad_kg: undefined }]],
    [
      'sólo el segundo item inválido',
      [{ ...ITEM_VALIDO, producto_id: 1 }, { ...ITEM_VALIDO, producto_id: 2, cantidad_kg: 0 }],
    ],
  ])('devuelve error de cantidad cuando hay %s', (_caso, carrito) => {
    // Act
    const errores = validarPedido(NOMBRE_VALIDO, CONTACTO_VALIDO, carrito);

    // Assert
    expect(errores).toContain('Todas las cantidades deben ser mayores a 0.');
  });

  it('no devuelve errores cuando el pedido es válido', () => {
    // Arrange
    const carrito = carritoValido();

    // Act
    const errores = validarPedido(NOMBRE_VALIDO, CONTACTO_VALIDO, carrito);

    // Assert
    expect(errores).toEqual([]);
  });
});

describe('agregarItem', () => {
  it('agrega un producto nuevo al final del carrito', () => {
    // Arrange
    const carrito = [{ ...ITEM_VALIDO }];
    const producto = { id: 2, nombre: 'Almendras', precio_por_kg: 200 };

    // Act
    const resultado = agregarItem(carrito, producto, 3);

    // Assert
    expect(resultado).toEqual([
      { ...ITEM_VALIDO },
      { producto_id: 2, nombre: 'Almendras', precio_por_kg: 200, cantidad_kg: 3 },
    ]);
  });

  it('suma la cantidad cuando el producto ya existe y no muta el carrito original', () => {
    // Arrange
    const carrito = [{ ...ITEM_VALIDO }];
    const copiaOriginal = structuredClone(carrito);
    const producto = { id: ITEM_VALIDO.producto_id, nombre: ITEM_VALIDO.nombre, precio_por_kg: ITEM_VALIDO.precio_por_kg };

    // Act
    const resultado = agregarItem(carrito, producto, 5);

    // Assert
    expect(resultado).not.toBe(carrito);
    expect(resultado).toEqual([{ ...ITEM_VALIDO, cantidad_kg: ITEM_VALIDO.cantidad_kg + 5 }]);
    expect(carrito).toEqual(copiaOriginal);
  });

  it('deja intactos los demás productos cuando suma la cantidad de uno', () => {
    // Arrange
    const otroItem = { producto_id: 2, nombre: 'Almendras', precio_por_kg: 200, cantidad_kg: 1 };
    const carrito = [{ ...ITEM_VALIDO }, otroItem];
    const producto = { id: ITEM_VALIDO.producto_id, nombre: ITEM_VALIDO.nombre, precio_por_kg: ITEM_VALIDO.precio_por_kg };

    // Act
    const resultado = agregarItem(carrito, producto, 5);

    // Assert
    expect(resultado[0].cantidad_kg).toBe(ITEM_VALIDO.cantidad_kg + 5);
    expect(resultado[1]).toEqual(otroItem);
  });
});

describe('totalCarrito', () => {
  it.each([
    ['carrito vacío', [], 0],
    ['un item', [{ cantidad_kg: 2, precio_por_kg: 100 }], 200],
    [
      'varios items',
      [
        { cantidad_kg: 2, precio_por_kg: 100 },
        { cantidad_kg: 1.5, precio_por_kg: 50 },
      ],
      275,
    ],
  ])('calcula el total con %s', (_caso, items, esperado) => {
    // Act
    const total = totalCarrito(items);

    // Assert
    expect(total).toBeCloseTo(esperado);
  });
});

describe('confirmarPedido', () => {
  it('no llama a crear cuando el pedido es inválido', async () => {
    // Arrange
    const crear = vi.fn();
    const carrito = [];

    // Act
    const resultado = await confirmarPedido({ clienteNombre: '', clienteContacto: '' }, carrito, crear);

    // Assert
    expect(crear).toHaveBeenCalledTimes(0);
    expect(resultado.errores).toEqual([
      'El nombre del cliente es requerido.',
      'El contacto del cliente es requerido.',
      'Agregá al menos un producto al pedido.',
    ]);
  });

  it('llama a crear una vez con el payload recortado cuando el pedido es válido', async () => {
    // Arrange
    const pedidoCreado = { id: 7, items: [{ producto_id: 1, cantidad_kg: 2 }] };
    const crear = vi.fn().mockResolvedValue(pedidoCreado);
    const carrito = [{ ...ITEM_VALIDO }];

    // Act
    const resultado = await confirmarPedido(
      { clienteNombre: '  Ana  ', clienteContacto: '  ana@mail.com  ' },
      carrito,
      crear,
    );

    // Assert
    expect(crear).toHaveBeenCalledTimes(1);
    expect(crear).toHaveBeenCalledWith({
      cliente_nombre: 'Ana',
      cliente_contacto: 'ana@mail.com',
      items: [{ producto_id: ITEM_VALIDO.producto_id, cantidad_kg: ITEM_VALIDO.cantidad_kg }],
    });
    expect(resultado).toEqual({ errores: [], pedido: pedidoCreado });
  });

  it('rechaza con el mismo error cuando crear falla', async () => {
    // Arrange
    const crear = vi.fn().mockRejectedValue(new Error('boom'));
    const carrito = carritoValido();

    // Act & Assert
    await expect(
      confirmarPedido({ clienteNombre: NOMBRE_VALIDO, clienteContacto: CONTACTO_VALIDO }, carrito, crear),
    ).rejects.toThrow('boom');
  });
});
