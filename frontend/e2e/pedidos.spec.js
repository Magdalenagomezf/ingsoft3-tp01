import { expect, test } from '@playwright/test'

// Cada test se arma SU producto, con nombre único y stock conocido, y lo borra
// al final: no depende de qué productos haya en QA ni deja nada atrás.
// Los pedidos a /api pasan por el front (baseURL), igual que los del navegador.
async function crearProducto (request, nombre, stockKg) {
  const r = await request.post('/api/productos', {
    data: { nombre, descripcion: 'creado por e2e', categoria: 'e2e', precio_por_kg: 1000, stock_kg: stockKg },
  })
  expect(r.status()).toBe(201)
  return await r.json()
}

async function borrarProducto (request, id) {
  const r = await request.delete(`/api/productos/${id}`)
  expect(r.status()).toBe(204)
}

// Lo que hace el cliente en la pantalla "Armar pedido"
async function armarPedido (page, producto, kg, cliente) {
  await page.goto('/armar-pedido')
  await page.getByLabel(`Cantidad en kg de ${producto.nombre}`).fill(String(kg))
  await page.getByRole('button', { name: `Agregar ${producto.nombre} al pedido` }).click()
  await page.getByLabel('Nombre del cliente').fill(cliente)
  await page.getByLabel('Contacto').fill('3510000000')
  await page.getByRole('button', { name: 'Confirmar pedido' }).click()
}

// "Pedidos" carga la lista después de abrir: esperamos a que termine
// (aparece un pedido o el mensaje de vacío) antes de afirmar que algo NO está.
async function esperarListaDePedidos (page) {
  await expect(page.getByText(/Pedido #\d+|Todavía no hay pedidos/).first()).toBeVisible()
}

test('armar un pedido lo muestra en Pedidos, y cancelarlo lo saca y devuelve el stock', async ({ page, request }) => {
  const ts = Date.now()                                          // único: no choca con otra corrida
  const producto = await crearProducto(request, `e2e producto ${ts}`, 10)
  const cliente = `e2e cliente ${ts}`

  await armarPedido(page, producto, 1, cliente)
  const confirmacion = page.getByText(/Pedido #\d+ confirmado/)
  await expect(confirmacion).toBeVisible()                       // espera sola: sin sleep
  const id = (await confirmacion.textContent()).match(/#(\d+)/)[1]

  await page.goto('/pedidos')
  await expect(page.getByText(cliente)).toBeVisible()            // el pedido que ESTE test creó
  await page.getByRole('button', { name: `Cancelar pedido #${id}` }).click()
  await expect(page.getByText(cliente)).toHaveCount(0)           // y la cancelación, comprobada

  const despues = await (await request.get(`/api/productos/${producto.id}`)).json()
  expect(despues.stock_kg).toBe(10)                              // el stock volvió entero

  await borrarProducto(request, producto.id)
})

test('pedir más kg que el stock muestra el error y no crea el pedido', async ({ page, request }) => {
  const ts = Date.now()
  const producto = await crearProducto(request, `e2e producto ${ts}`, 1)
  const cliente = `e2e cliente ${ts}`

  await armarPedido(page, producto, 5, cliente)                  // 5 kg de un producto con 1 kg
  await expect(page.getByRole('alert')).toContainText('stock insuficiente')   // el usuario ve el error…

  await page.goto('/pedidos')
  await esperarListaDePedidos(page)
  await expect(page.getByText(cliente)).toHaveCount(0)           // …y no se creó ningún pedido

  await borrarProducto(request, producto.id)
})

test('un pedido confirmado sigue en Pedidos al recargar, y cancelado no vuelve', async ({ page, request }) => {
  const ts = Date.now()
  const producto = await crearProducto(request, `e2e producto ${ts}`, 10)
  const cliente = `e2e cliente ${ts} persiste`

  await armarPedido(page, producto, 2, cliente)
  const confirmacion = page.getByText(/Pedido #\d+ confirmado/)
  await expect(confirmacion).toBeVisible()
  const id = (await confirmacion.textContent()).match(/#(\d+)/)[1]

  await page.goto('/pedidos')
  await expect(page.getByText(cliente)).toBeVisible()
  await page.reload()                                            // lo que se ve ahora lo trae la API desde la base
  await expect(page.getByText(cliente)).toBeVisible()
  await page.getByRole('button', { name: `Cancelar pedido #${id}` }).click()
  await expect(page.getByText(cliente)).toHaveCount(0)           // la cancelación se vio en pantalla
  await page.reload()
  await esperarListaDePedidos(page)
  await expect(page.getByText(cliente)).toHaveCount(0)           // y también llegó a la base

  await borrarProducto(request, producto.id)
})