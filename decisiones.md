# Decisiones — Proyecto IngSoft3

## TP1 — Control de versiones

### Por qué Git no pudo resolver el conflicto solo
Las ramas `feature/titulo-a` y `feature/titulo-b` partieron las dos del mismo commit de `main`, 
y ambas modificaron la misma línea del `README.md`, cada una con un texto distinto
("versión A" vs "versión B"). El merge de tres vías que usa Git compara el commit común (la base),
la rama A y la rama B — y solo puede fusionar automático cuando los cambios no se pisan. Como las
dos tocaron exactamente la misma línea con contenido distinto, Git no tiene forma de decidir y necesita que una persona lo resuelva.

La rama B tendría que haber traído los cambios de `main` (ya con A mergeada) antes de abrir su
 propio PR — así el conflicto aparecía localmente, antes de llegar a GitHub, sabiendo de antemano qué había en A.

### Problemas encontrados y cómo los resolví
Para que el conflicto entre las ramas A y B apareciera de verdad (y no se mergeara solo), las dos
  tenían que partir del mismo commit de `main` y modificar la misma línea del README. Si una rama
  partía de la otra en vez de `main`, Git no generaba ningún conflicto — hubo que prestar atención
  al orden: crear ambas ramas desde `main`, y recién mergear la primera después de tener las dos
  abiertas.

### Uso de IA
Usé Claude como guía durante el TP1: para entender la protección de rama, interpretar los
marcadores de conflicto antes de resolverlos, redactar la descripción de la release, y resolver
dudas puntuales. Verifiqué cada paso ejecutándolo yo misma y mirando el resultado real en GitHub.

## TP2 — Contenedores

### Qué app elegí y por qué
Mallki Nueces: un ecommerce mayorista de nueces para el trabajo de mi papá. Es una idea que ya tenia pendiente y decidi aprovechar la oportunidad de hacerla, ademas cumple los criterios de la guía: backend con API (Go), frontend SPA (React + Vite), 
base de datos relacional (PostgreSQL), corre local sin nada raro, y el tamaño es chico a propósito — 
catálogo, armado de pedido y listado
de pedidos, sin login ni pagos (eso queda para más adelante, fuera del alcance de este TP).

### Decisiones de contenerización
- **Imágenes base**: `golang:1.24-alpine` para compilar el backend (coincide con la versión de mi
  `go.mod`), y `alpine:3.20` para correrlo — como el binario de Go es estático
  (`CGO_ENABLED=0`), no necesita ningún runtime de Go instalado en la imagen final. Para el
  frontend, `node:22-alpine` compila el build de Vite, y `nginx:alpine` sirve los estáticos
  resultantes.
- **Multi-stage en los dos**: cada Dockerfile tiene una etapa de compilación (pesada, con el SDK/
  toolchain) y una etapa final mínima que solo recibe el resultado ya compilado —
  `golang:1.24-alpine` pesa 83.5MB, mi imagen final del backend, 9.11MB.
- **Qué persiste y qué no**: solo los datos de Postgres, en el volumen nombrado `db_data`. Backend
  y frontend son efímeros a propósito — su código vive copiado *dentro* de la imagen (no montado
  como volumen), así que cualquier cambio requiere reconstruir la imagen, no editar un archivo
  vivo. Confirmé esto con la prueba de `down`/`up` vs `down -v` en `evidencias.md`.
- **Red y nombres**: dentro de compose, el backend le habla a la base por el nombre del servicio
  (`db:5432`, el puerto interno de Postgres) — no por `localhost` ni por el puerto publicado hacia
  afuera, que es un número distinto y solo le sirve a mi máquina, no a otros contenedores.

### Problemas encontrados y cómo los resolví
- Al probar `docker-compose.registry.yml` (que baja las imágenes en vez de compilarlas), corrimos
  `docker builder prune -af` para forzar una descarga limpia de verdad — eso también me borró las
  imágenes base (`golang:1.24-alpine`, `node:22-alpine`, etc.) que necesitaba para la comparación de
  tamaños de `evidencias.md`. Tuve que volver a bajarlas con `docker pull` una por una.

### Uso de IA
Usé Claude Code para escribir el scaffolding inicial del backend (arquitectura en capas) y del
frontend, a partir de un prompt donde definí el alcance, el modelo de datos y las restricciones
explícitas (sin pagos, sin login, sin Docker generado por la IA). Usé Claude para guiarme paso a
paso escribiendo los Dockerfiles, el `docker-compose.yml` y el `docker-compose.registry.yml` — pero
cada comando lo corrí yo misma, y verifiqué el resultado real en mi terminal y en el navegador. Cuando algo no
coincidía con lo esperado, lo diagnostiqué con mis propios comandos antes de pedir ayuda para interpretarlo.

## TP3 — Planificación

**Duración del sprint:** 1 semana. Elegí ese numero porque coincide con el ritmo real
de entrega de la materia (un TP por clase), alinear el sprint con eso hace que cada clase sea el cierre de un sprint, y refleja lo que de verdad estoy entregando, no un número arbitrario.

**Limite de trabajo en progreso:** 2. Es la regla de arranque de la guia (cantidad de personas + 1);
trabajando sola, 1 + 1 = 2. La herramienta no bloquea que sigas agregando tarjetas, pero el contador se pone en rojo cuando te pasás del límite — así el problema se nota en vez de acumularse en silencio.
Si con el tiempo, nunca alcanzo ese limite, esta demasiado alto y debo bajarlo.

**Diagnostico de la historia mal escrita:** 
"Como desarrollador quiero crear la tabla
usuarios" mezcla una tarea tecnica (crear una tabla) con el formato de historia.Nadie quiere una tabla en la base de datos, es un detalle de implementacion, no un requisito para alguien. Le falta un beneficio real- el "para que" le importa a un usuario, no a la base de datos.
Cómo la reescribiría: "Como usuario quiero ver mi perfil con los datos que cargué antes, para no tener que completarlos de nuevo cada vez que entro" — con un criterio verificable (por ejemplo: los datos del perfil siguen ahí después de cerrar sesión y volver a entrar). "Crear la tabla usuarios" pasaría a ser una **tarea** técnica dentro de esa historia.

**Problemas que encontre:**
- Al token de gh le faltaba el scope 'project'; se resolvio con `gh auth refresh -s project`.
- cmd.exe no interpreta comillas simples como GitHub CLI espera (ej: `--owner '@me'` tiraba "unknown owner type"); hubo que sacar las comillas o cambiarlas por dobles.
- Se me subio `ci.yml` vacio por no confirmar el contenido antes
  de commitear — lo solucione revisando el diff en "Files changed" del PR antes de  mergear, en vez de mergear a ciegas.
- Las dos tareas quedaron colgando de la épica en vez de la historia — al agregarlas por la web, seguí parada en la página de la épica en vez de ir   a la de la historia (#7).
  Lo noté comparando el tablero contra el esquema de la Clase 3 (épica → historia → tarea)  y lo corregí con `gh issue edit 8 --parent 7` y `gh issue edit 9 --parent 7`.

**Uso de IA:** Usé la IA (Claude) para guiarme paso a paso en toda la configuración de GitHub Projects, traduciendo los comandos del video del profesor a cmd.exe de Windows, para pensar variantes de la historia mal escrita hasta llegar a una que cumpliera los cuatro criterios, y para diagnosticar dos problemas reales: el token de `gh` sin el scope `project`, y el archivo `ci.yml` que se subió vacío dos veces seguidas por no guardarlo antes de comitear.
Las decisiones (duración del sprint, límite de WIP) las tomé yo. Verifiqué cada paso mirando el estado real en GitHub.

## TP4 — CI: Pipelines as Code

### Estructura elegida del pipeline
Dos jobs, `build-backend` y `build-frontend`, corriendo en paralelo. La separación no es arbitraria: coincide con que desde el TP2 tengo dos Dockerfiles distintos (`backend/Dockerfile` y `frontend/Dockerfile`), así que el pipeline refleja cómo está armada la app de verdad. Van en paralelo porque son builds independientes — uno no necesita el resultado del otro, así que correrlos en serie solo sumaría tiempo de espera sin ninguna ganancia.

### Qué cachea y qué pasa si desaparece
El cache guarda las capas de Docker de cada Dockerfile, en particular las más caras: la descarga de dependencias de Go (`go mod download`) en el backend y la instalación de paquetes (`npm ci`) en el frontend. Lo probé con un commit vacío entre dos corridas del mismo PR, y las 7 capas del backend y las 6 del frontend salieron `CACHED` por completo. Cada job usa su propio `scope` (`backend` y `frontend`) para que no se pisen el cache entre sí — si compartieran uno solo, el último job en terminar sobrescribiría el cache del otro. Si el cache desaparece en algún momento (la plataforma lo puede desalojar), el pipeline sigue funcionando exactamente igual, solo que más lento — no hay ninguna dependencia funcional de que exista, es pura optimización de tiempo.

### Por qué el pipeline construye con mi Dockerfile en vez de compilar por su cuenta
El workflow no tiene ni una línea de Go ni de npm — usa exactamente los mismos `backend/Dockerfile` y `frontend/Dockerfile` que ya tenía del TP2. Si el pipeline compilara por su cuenta (con `go build` o `npm run build` directo en el YAML), tendría dos definiciones distintas de cómo se construye la app: la que usa el pipeline para verificar, y la que después uso para desplegar. Tarde o temprano esas dos definiciones divergen, y terminaría verificando algo que no es lo que realmente corre.

### Problemas encontrados y cómo los resolví
- Para la demostración de romper el build a propósito, agregué un import de un paquete que no existe en `main.go`. VS Code tiene activado el formateador automático de Go (`goimports`), que borra los imports no usados apenas guardás el archivo — así que mi import "roto" desaparecía solo antes de poder commitearlo. Lo resolví usando un import en blanco (`_ "nueces-backend/noexiste"`), que el formateador no toca porque está marcado explícitamente como intencional.
- Al mergear con squash el PR que rompía y arreglaba el build (dos commits en la misma rama), la pestaña "Files changed" mostró "No changes to show" — al principio pensé que algo se había perdido, pero es esperable: el resultado neto entre los dos commits, comparado contra `main`, es cero cambios (agregué una línea y la misma línea la saqué después).

### Uso de IA
Usé Claude para traducir el video y la guía del profesor (escritos sobre .NET) a mi stack en Go, y para guiarme paso a paso. También me ayudó a diagnosticar el problema del formateador de Go que borraba el import roto. Cada paso lo corrí yo misma y verifiqué el resultado real en GitHub (los checks, el cache en el log, el badge en el README) antes de seguir.

## TP5 — Calidad automatizada: tests, coverage y umbral

> 🔗 **Links pendientes** (se completan cuando el pipeline corra los tests):
> - Corrida con el resumen de cobertura y el reporte descargable: `🔗 PENDIENTE …/actions/runs/<id>`
> - Corrida **roja por umbral**, con el número en el log: `🔗 PENDIENTE …/actions/runs/<id>`
> - PR #1 (rojo → tests → verde → merge): `🔗 PENDIENTE …/pull/<n>`
> - PR #2 (queda **abierto y en rojo** hasta la defensa): `🔗 PENDIENTE …/pull/<m>`

### Qué dejé afuera de la cuenta de cobertura, y por qué

**Backend (Go).** Excluyo por **nombre de archivo** —digo qué se saca, no qué entra—, así un
archivo nuevo entra solo a la cuenta: si no tiene tests, el número baja y el umbral me avisa. Con
una lista de lo que entra pasaría lo contrario: lo que me olvide de nombrar desaparecería de la
medición sin avisar.

| Excluido | Por qué |
|---|---|
| `main.go` | Arranque: arma los repositorios, servicios y rutas. No tiene reglas; si está mal, la app no levanta |
| `db.go` | Arranque: lee `DATABASE_URL`, conecta y crea las tablas. Configuración, no lógica |
| `*/model.go` | Structs con campos y tags JSON, sin ningún comportamiento |
| `*/repository.go` | SQL contra Postgres. Su lugar natural es un test de integración contra una base real, no un unit test. Igual lo ejecutan los tests de sqlmock, pero no lo cuento |

Lo que **no** excluí, a propósito: `handler.go` y `internal/httpx`. Los handlers tienen
comportamiento —un id no numérico da 400, `ErrNotFound` da 404, un error de validación da 400 con
su mensaje, cualquier otro da 500—, y si ese mapeo se rompe la API contesta mal. Excluirlos
hubiera sido esconder código que no testeé, así que les escribí tests con `httptest`. `httpx`
tiene poca lógica (`if body != nil`, `IsNoRows`) pero tiene, y los tests de los handlers la cubren.

Medición: **62,2 %** midiendo todo, **94,0 %** (158 de 168 sentencias) sacando lo de la tabla.
Mido con `-coverpkg=./...`: sin eso, Go sólo cuenta lo que ejecutan los tests del mismo paquete, y
`httpx` daría 0 % aunque los handlers lo usen en cada respuesta.

**Frontend.** `include: ['src/lib/**']` en `vite.config.js`: ahí vive la lógica pura del pedido.
Quedan afuera los componentes de React: testearlos pide jsdom y Testing Library, y la pantalla
completa se verifica end-to-end en el TP7. Desde vitest 4 el `include` es obligatorio: sin él,
vitest mide sólo los archivos que los tests importan, y un archivo nuevo sin tests ni aparecería.
Medición: **100 % de líneas (19/19) y 100 % de ramas (14/14)**.

### Umbral de cobertura

**Frontend: `lines: 90, branches: 90`** — las dos métricas, porque la de líneas sola es la que
más miente. Mi medición real es 100 / 100, pero la base es chica (14 ramas) y cada rama pesa
~7 %: con una rama sin cubrir quedo en 92,9 % y pasa; con dos, 85,7 % y frena; con tres, 78,6 %.
Con 90 tolero un descuido, pero no una función nueva con varios caminos sin tests. Con 80 dejaría
pasar dos ramas sin probar sobre una base de 14; con 100, cualquier refactor chico rompería el
build y terminaría apagando el umbral.

**Backend: 85 % de sentencias**, sobre el total (no por paquete). Hoy mide **94,0 %: 158 de 168
sentencias**, así que cada sentencia pesa ~0,6 %. Con 85 el número me frena si se pierden más de
15 sentencias cubiertas, o si entra un archivo nuevo de más de ~17 sentencias sin ningún test (que
es justo lo que quiero atrapar: un handler o servicio nuevo sin tests). Con 90 el margen sería de 6
sentencias, y en Go cada `if err != nil { return … }` es una sentencia: varias de las 10 que hoy no
cubro son ramas de error que un unit test no alcanza (como el 500 inalcanzable de `pedido.Create`),
y cualquier refactor que sume dos o tres de ésas rompería el build y terminaría apagando el umbral.
Para subirlo a 90 tendría que cubrir antes las ramas de error de los handlers que hoy quedan afuera.

Go **no mide cobertura de rama**, sólo de sentencias (`go tool cover` no tiene la métrica): el
número del backend es de sentencias y lo reporto así.

### El ejercicio de la rama sin cubrir

- **Qué línea:** `frontend/src/lib/pedido.js:28`, el `: item` del ternario dentro del `map` de
  `agregarItem`. El reporte de v8 la marcaba en *Uncovered Line #s* con 13/14 ramas.
- **Qué entrada la recorre:** un carrito con **dos** productos (ids 1 y 2) al que le agrego el 1.
  El test existente usaba un carrito de un solo item, así que el `map` nunca encontraba uno que
  *no* coincidiera.
- **Qué decidí:** agregar el test *«deja intactos los demás productos cuando suma la cantidad de
  uno»*, que verifica que el producto 2 vuelve sin cambios. Las ramas
  pasaron a 14/14.

Y una del backend que decidí **no** cubrir: `backend/internal/pedido/handler.go`, el `500` genérico
de `Create`. Es inalcanzable: `Service.Create` siempre devuelve un `*Error`, nunca un error
cualquiera. No hay entrada que lo recorra; lo que correspondería es simplificar el handler, no
agregar un test.

### Mi stack contra la tabla «Tu stack, de un vistazo»

| Lo que hay que lograr | Backend (Go) | Frontend (JS) |
|---|---|---|
| Dónde viven los tests | `*_test.go` al lado del código, mismo paquete | `algo.test.js` al lado del código |
| Test parametrizado | table-driven con `t.Run` | `it.each` |
| Que la dependencia entre desde afuera | interfaz `producto.Repo` recibida en `NewService` | el cliente de la API entra por parámetro (`crear` en `confirmarPedido`) |
| Fabricar el doble | mock escrito a mano (`mockRepo`, cuenta llamadas); `go-sqlmock` para el `*sql.DB` | `vi.fn()` |
| Medir la cobertura | `go test -coverpkg=./... -coverprofile=coverage.out` | `vitest run --coverage` (`@vitest/coverage-v8`) |
| Umbral que rompe el build | Go no tiene bandera: `backend/scripts/coverage.sh` lee el `total:` de `go tool cover -func`, lo compara con `UMBRAL=85` y sale con error si no llega (también si falla un test o si mide 0 archivos). Es el mismo script en mi máquina y en la etapa `test` del Dockerfile | `coverage.thresholds` de vitest |
| Qué entra en la cuenta | filtro del perfil que excluye por archivo | `include` de `coverage` |
| Reporte legible | `go tool cover -html` | reporters `html` y `lcov` |
| Herramientas en la etapa de tests del Dockerfile | `FROM build AS test`: la imagen `golang` ya trae `go test` y `go tool cover`, y el `go mod download` del build ya bajó `go-sqlmock` (está en `go.mod`), así que los tests corren sin red. La imagen final (`alpine`) sólo copia el binario | `FROM build AS test` sobre un `npm ci` sin `--omit=dev`: vitest y `@vitest/coverage-v8` son devDependencies y entran a la etapa. nginx sólo copia `dist/` |

### Problemas encontrados y cómo los resolví

- **Nombres con sólo espacios pasaban la validación.** Escribiendo los tests vi que `validate`
  rechazaba `""` pero aceptaba `"   "`, igual que `cliente_nombre` y `cliente_contacto` en
  pedidos. Lo arreglé test primero: convertí esos tests en parametrizados con `""`, `"   "` y
  `"\t"`, los vi en rojo, y recién después agregué `strings.TrimSpace`.
- **Errores de validación que quedaban pegados en la pantalla.** Al sacar la lógica del pedido a
  funciones puras, la validación pasó a correr después del envío; si el backend fallaba, seguían
  en pantalla los errores del intento anterior junto al error del servidor. Los tests no lo veían
  porque prueban la lógica, no el componente. Lo resolví dejando que el componente valide primero
  con `validarPedido`, como antes, y `confirmarPedido` conserva su propia validación como guardia.
- **Un mutante de la cantidad que ningún test podía matar.** Cambiar `cantidad_kg <= 0` por
  `< 0` no ponía nada en rojo: la condición empezaba con `!item.cantidad_kg`, que ya rechaza el 0,
  así que el `<= 0` nunca decidía nada en el borde. No faltaba un test, sobraba código: la
  simplifiqué a `!(item.cantidad_kg > 0)` y ahora ese mutante lo matan dos tests.
- **El orden de un `map` con sqlmock.** `pedido.Create` recorre un `map` para descontar stock, y
  en Go ese orden es aleatorio. Con sqlmock, que espera las queries en orden, un test con dos
  productos distintos fallaría a veces. Uso un solo producto por test (con dos items del mismo,
  para probar que suma las cantidades).
- **Sentencias contadas dos veces con `-coverpkg`.** Con `-coverpkg=./...` cada paquete de tests
  escribe todos los bloques en el perfil, así que el mismo bloque aparece una vez por paquete.
  `go tool cover` los fusiona y el 94,0 % sale bien, pero al sumar sentencias a mano me daba
  «163 de 342». El script cuenta cada bloque una sola vez (cubierto si algún paquete lo ejecutó):
  158 de 168, igual que la medición local.
- **El script y los finales de línea de Windows.** Con `core.autocrlf` el `.sh` se guardaría con
  CRLF en mi máquina y el `sh` del contenedor Linux no lo correría. Lo fijé con `.gitattributes`
  (`*.sh text eol=lf`).
- **vitest 5.** Tengo Vite 8, y vitest 5.0.2 es la versión que lo soporta. `@vitest/coverage-v8`
  tiene que ser exactamente la misma versión (5.0.2). Y desde vitest 4 el `include` de la
  cobertura es obligatorio (ver arriba).
- **Windows Defender bloqueando el repo.** Git fallaba al crear la rama con *«cannot lock ref …
  No such file or directory»*, y después `npm test` y `npm run build` fallaban con `ENOENT` en
  `node_modules/.vite-temp`. Era el *Acceso controlado a carpetas*: mi carpeta Documentos está en
  `D:\Magdi\documentos` y queda protegida. Lo confirmé en el registro de eventos de Defender (evento
  1123, proceso `node.exe`) y lo resolví permitiendo `git.exe`, `bash.exe` y `node.exe`, sin
  apagar la protección.

### Pendiente de escribir

- Qué lógica elegí testear y por qué ésa.
- Por qué coverage alto no garantiza calidad, con mi ejemplo.
- El Pull Request bloqueado: qué check, en qué métrica y qué escribí para arreglarlo.
- El refactor para poder mockear (`producto.Service` pasó de `*Repository` a la interfaz `Repo`).
- Uso de IA.