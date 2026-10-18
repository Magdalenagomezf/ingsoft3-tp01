import { expect, test } from '@playwright/test'

// La api tiene su propia dirección: NO es la del front. Sin barra al final.
const API = process.env.API_BASE_URL || 'http://localhost:8080'

// ayudas, para no repetir: piden la lista y la devuelven ya convertida
async function listarProductos (request) {
  const r = await request.get(`${API}/api/productos`)
  expect(r.status()).toBe(200)
  return await r.json()
}

async function listarPedidos (request) {
  const r = await request.get(`${API}/api/pedidos`)
  expect(r.status()).toBe(200)
  return await r.json()
}

test('el alta de un producto guarda en la base de verdad, y el borrado lo saca', async ({ request }) => {
  const nombre = `api ${Date.now()}`                            // único: no choca con otra corrida
  const alta = await request.post(`${API}/api/productos`, {
    data: { nombre, descripcion: 'creado por integracion', categoria: 'api', precio_por_kg: 1000, stock_kg: 10 },
  })
  expect(alta.status()).toBe(201)                               // la api dice «creado»
  const creado = await alta.json()

  const lista = await listarProductos(request)                  // y la BASE lo devuelve: no es un doble
  expect(lista.some(p => p.nombre === nombre)).toBe(true)

  const borrado = await request.delete(`${API}/api/productos/${creado.id}`)
  expect(borrado.status()).toBe(204)                            // «borrado, no hay nada que devolver»

  const despues = await listarProductos(request)                // y el borrado, comprobado
  expect(despues.some(p => p.nombre === nombre)).toBe(false)
})

test('un producto con nombre vacío lo rechaza la api, y no crea nada', async ({ request }) => {
  const antes = await listarProductos(request)

  const alta = await request.post(`${API}/api/productos`, {
    data: { nombre: '', precio_por_kg: 1000, stock_kg: 10 },   // todo válido salvo el nombre
  })
  expect(alta.status()).toBe(400)                               // el contrato del error

  const despues = await listarProductos(request)
  expect(despues.length).toBe(antes.length)                     // y de verdad no se creó nada
})

test('un pedido con más kg que el stock lo rechaza la api, sin tocar stock ni pedidos', async ({ request }) => {
  const nombre = `api ${Date.now()} stock`
  const alta = await request.post(`${API}/api/productos`, {
    data: { nombre, descripcion: 'creado por integracion', categoria: 'api', precio_por_kg: 1000, stock_kg: 1 },
  })
  expect(alta.status()).toBe(201)
  const producto = await alta.json()
  const pedidosAntes = await listarPedidos(request)

  const pedido = await request.post(`${API}/api/pedidos`, {
    data: { cliente_nombre: `api ${Date.now()}`, cliente_contacto: '3510000000',
            items: [{ producto_id: producto.id, cantidad_kg: 5 }] },   // 5 kg de un producto con 1 kg
  })
  expect(pedido.status()).toBe(400)
  expect((await pedido.json()).error).toContain('stock insuficiente')

  const releido = await (await request.get(`${API}/api/productos/${producto.id}`)).json()
  expect(releido.stock_kg).toBe(1)                              // la transacción no descontó nada
  const pedidosDespues = await listarPedidos(request)
  expect(pedidosDespues.length).toBe(pedidosAntes.length)       // ni creó un pedido

  const borrado = await request.delete(`${API}/api/productos/${producto.id}`)
  expect(borrado.status()).toBe(204)                            // limpia lo que creó
})