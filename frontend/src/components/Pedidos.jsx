import { useEffect, useState } from 'react';
import { deletePedido, getPedidos } from '../api';
import { PageBanner } from './PageBanner';
import bannerFoto from '../assets/fotos/mitades-luz.jpg';

function formatearFecha(fechaISO) {
  try {
    return new Date(fechaISO).toLocaleString('es-AR', {
      dateStyle: 'short',
      timeStyle: 'short',
    });
  } catch {
    return fechaISO;
  }
}

const ESTADO_BADGE_CLASSES = {
  pendiente: 'bg-yellow-100 text-yellow-800',
};
const ESTADO_BADGE_DEFAULT = 'bg-secondary-100 text-secondary-700';

export function Pedidos() {
  const [pedidos, setPedidos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [cancelando, setCancelando] = useState(null);
  const [errorCancelar, setErrorCancelar] = useState(null);

  useEffect(() => {
    async function cargar() {
      setLoading(true);
      setError(null);
      try {
        const data = await getPedidos();
        setPedidos(data ?? []);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    cargar();
  }, []);

  async function cancelar(id) {
    setCancelando(id);
    setErrorCancelar(null);
    try {
      await deletePedido(id);
      setPedidos((actuales) => actuales.filter((pedido) => pedido.id !== id));
    } catch (err) {
      setErrorCancelar(err.message);
    } finally {
      setCancelando(null);
    }
  }

  return (
    <div className="flex flex-col">
      <PageBanner titulo="Pedidos" foto={bannerFoto} alt="Mitades de nuez en primer plano" />

      <div className="mx-auto w-full max-w-5xl px-4 py-10 sm:px-6">
        {loading && <p className="text-secondary-500">Cargando pedidos...</p>}
        {error && (
          <p className="rounded-md bg-red-50 px-4 py-3 text-red-700">Error al cargar pedidos: {error}</p>
        )}
        {errorCancelar && (
          <p role="alert" className="mb-4 rounded-md bg-red-50 px-4 py-3 text-red-700">
            No se pudo cancelar el pedido: {errorCancelar}
          </p>
        )}
        {!loading && !error && pedidos.length === 0 && (
          <p className="text-secondary-500">Todavía no hay pedidos registrados.</p>
        )}

        {!loading && !error && pedidos.length > 0 && (
          <div className="flex flex-col gap-4">
            {pedidos.map((pedido) => (
              <article
                key={pedido.id}
                className="rounded-2xl bg-white p-5 shadow-sm ring-1 ring-secondary-100 transition-shadow hover:shadow-md"
              >
                <header className="flex items-center justify-between">
                  <h3 className="text-lg font-bold text-secondary-800">Pedido #{pedido.id}</h3>
                  <div className="flex items-center gap-3">
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wide ${
                        ESTADO_BADGE_CLASSES[pedido.estado] ?? ESTADO_BADGE_DEFAULT
                      }`}
                    >
                      {pedido.estado}
                    </span>
                    <button
                      type="button"
                      aria-label={`Cancelar pedido #${pedido.id}`}
                      onClick={() => cancelar(pedido.id)}
                      disabled={cancelando === pedido.id}
                      className="rounded-md border border-red-600 px-2 py-1 text-xs font-semibold text-red-600 transition-colors hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {cancelando === pedido.id ? 'Cancelando...' : 'Cancelar'}
                    </button>
                  </div>
                </header>
                <p className="mt-1 text-secondary-700">
                  <strong>{pedido.cliente_nombre}</strong> — {pedido.cliente_contacto}
                </p>
                <p className="mt-0.5 text-sm text-secondary-500">{formatearFecha(pedido.fecha_creacion)}</p>
                <div className="mt-3 overflow-x-auto">
                  <table className="w-full border-collapse text-sm">
                    <thead>
                      <tr>
                        <th className="border-b border-secondary-200 px-2 py-2 text-left">Producto</th>
                        <th className="border-b border-secondary-200 px-2 py-2 text-left">Cantidad (kg)</th>
                        <th className="border-b border-secondary-200 px-2 py-2 text-left">Precio/kg</th>
                      </tr>
                    </thead>
                    <tbody>
                      {pedido.items.map((item) => (
                        <tr key={item.id}>
                          <td className="border-b border-secondary-100 px-2 py-2">{item.producto_nombre}</td>
                          <td className="border-b border-secondary-100 px-2 py-2">{item.cantidad_kg}</td>
                          <td className="border-b border-secondary-100 px-2 py-2">
                            ${item.precio_unitario.toFixed(2)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
