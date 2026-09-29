// Lógica pura del pedido: validación, carrito y confirmación.
// Sin imports de React para poder testear sin renderizar componentes.

/**
 * Valida los datos del cliente y el carrito antes de confirmar un pedido.
 * @returns {string[]} lista de mensajes de error (vacía si todo es válido)
 */
export function validarPedido(clienteNombre, clienteContacto, carrito) {
  const errores = [];
  if (!clienteNombre.trim()) errores.push('El nombre del cliente es requerido.');
  if (!clienteContacto.trim()) errores.push('El contacto del cliente es requerido.');
  if (carrito.length === 0) errores.push('Agregá al menos un producto al pedido.');
  if (carrito.some((item) => !(item.cantidad_kg > 0))) {
    errores.push('Todas las cantidades deben ser mayores a 0.');
  }
  return errores;
}

/**
 * Agrega un producto al carrito sin mutar el array ni los items originales.
 * Si el producto ya está en el carrito, suma la cantidad al item existente.
 * @returns {Array} nuevo carrito
 */
export function agregarItem(carrito, producto, cantidadKg) {
  const existente = carrito.find((item) => item.producto_id === producto.id);
  if (existente) {
    return carrito.map((item) =>
      item.producto_id === producto.id
        ? { ...item, cantidad_kg: item.cantidad_kg + cantidadKg }
        : item,
    );
  }
  return [
    ...carrito,
    {
      producto_id: producto.id,
      nombre: producto.nombre,
      precio_por_kg: producto.precio_por_kg,
      cantidad_kg: cantidadKg,
    },
  ];
}

/**
 * Calcula el total del carrito sumando cantidad_kg * precio_por_kg de cada item.
 */
export function totalCarrito(items) {
  return items.reduce((acc, item) => acc + item.cantidad_kg * item.precio_por_kg, 0);
}

/**
 * Valida y, si no hay errores, confirma el pedido llamando a `crear`.
 * @param {{clienteNombre: string, clienteContacto: string}} datosCliente
 * @param {Array} carrito
 * @param {(payload: object) => Promise<object>} crear
 * @returns {Promise<{errores: string[], pedido: object|null}>}
 */
export async function confirmarPedido({ clienteNombre, clienteContacto }, carrito, crear) {
  const errores = validarPedido(clienteNombre, clienteContacto, carrito);
  if (errores.length > 0) {
    return { errores, pedido: null };
  }

  const pedido = await crear({
    cliente_nombre: clienteNombre.trim(),
    cliente_contacto: clienteContacto.trim(),
    items: carrito.map(({ producto_id, cantidad_kg }) => ({ producto_id, cantidad_kg })),
  });

  return { errores: [], pedido };
}
