# Decisiones — Proyecto IngSoft3

🔗 **Enlaces del TP7** (la sección del TP7 está al final del documento)
- Paquete del backend: https://github.com/users/Magdalenagomezf/packages/container/package/ingsoft3-tp01-backend
- Paquete del frontend: https://github.com/users/Magdalenagomezf/packages/container/package/ingsoft3-tp01-frontend
- Cada versión de los dos paquetes está etiquetada `sha-<commit>`, con los 40 caracteres. La de PROD es `sha-fc3fdfeb7059c0a1d454f809257ad096afbf54ac`.
- Integración en verde y e2e en rojo, frenando la promoción:
- Commit que rompió la app: `681cb30`, *fix: createPedido apunta a /pedido* (PR #33)
- Corrida: https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/39181297975
- Reportes de esa corrida (en *Artifacts*, al pie del Summary): `playwright-report-integracion` (verde) y `playwright-report-e2e` (rojo)
- Corrida completa en verde hasta PROD, con el arreglo (PR #34, `fc3fdfe`):https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/39250463867 
- Release `v7.0.0`: https://github.com/Magdalenagomezf/ingsoft3-tp01/releases/tag/v7.0.0
- QA: https://mallki-front-qa.onrender.com (api: https://mallki-api-qa.onrender.com)
- PROD: https://mallki-front-prod.onrender.com (api: https://mallki-api-prod.onrender.com)

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

> 🔗 **Evidencias**
> - Resumen de cobertura y reporte descargable (artefactos `coverage-backend` y `coverage-frontend`), corrida verde final del PR #20 (commit `2528daf`): https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/37497868894
> - Corrida **roja por umbral del frontend** (PR #20, commit `99e1a15`): https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/37497621878
> - Corrida **roja por umbral del backend** (PR #21, commit `c2b951d`): https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/37502568352
> - PR 1 (rojo → tests → verde → merge): https://github.com/Magdalenagomezf/ingsoft3-tp01/pull/20
> - PR 2 (queda **abierto y en rojo** hasta la defensa): https://github.com/Magdalenagomezf/ingsoft3-tp01/pull/21

### Qué lógica elegí testear y por qué ésa

Empecé por donde un bug le cuesta plata a Mallki. El peor es **vender nuez que no hay**: si un
pedido pasa sin stock, le prometí mercadería a un cliente y no la tengo. Por eso testeé primero
`pedido.Service.Create` con sqlmock: que rechace el pedido cuando falta stock o el producto no
existe (y haga rollback), y que en el camino feliz descuente el stock y confirme la transacción.

Después, las **validaciones de pedido y producto**: nombre o contacto vacíos o con sólo espacios,
precio cero o negativo, stock negativo, pedido sin items, cantidades no positivas. Son la puerta
de entrada: un dato malo que pasa acá termina guardado y usado en todo lo demás.

Y los **códigos HTTP de los handlers**: id no numérico → 400, no encontrado → 404, validación →
400 con el mensaje del servicio, falla del repositorio → 500. El frontend decide qué mostrarle al
cliente según ese código, así que si el mapeo se rompe la pantalla dice algo equivocado.

En el frontend testeé la lógica del pedido (`validarPedido`, `agregarItem`, `totalCarrito`,
`confirmarPedido`): el total del carrito es lo que el cliente va a pagar.

### El refactor para poder mockear

Antes, `producto.Service` guardaba un `*Repository`, el struct concreto que habla con Postgres por
`*sql.DB`. No había forma de darle otra cosa: para testear la validación o el 404 tenía que
levantar una base real. Lo cambié para que dependa de la interfaz `Repo` (los métodos que el
servicio usa: listar, buscar, crear, actualizar, borrar), que `NewService` recibe por parámetro.
En los tests le paso `mockRepo`, un doble escrito a mano que devuelve lo que necesito y cuenta las
llamadas (así pruebo, por ejemplo, que un producto inválido nunca llega al repositorio).

`main.go` no cambió: en Go las interfaces se cumplen solas, sin declararlo. `*Repository` ya tiene
esos métodos, así que `producto.NewService(productoRepo)` sigue compilando igual.

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

**Backend: 85 % de sentencias**, sobre el total. Hoy mide **94,0 % (158 de 168)**: con 85 me frena
un archivo nuevo de más de ~17 sentencias sin tests, que es justo lo que quiero atrapar. No puse 90
porque el margen sería de 6 sentencias, y varias de las 10 que no cubro son ramas de error que un
unit test no alcanza: cualquier refactor chico rompería el build y terminaría apagando el umbral.

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

### Por qué coverage alto no garantiza calidad

Mi ejemplo está en `validarPedido`, en `frontend/src/lib/pedido.js`. La condición de cantidad era
`!item.cantidad_kg || item.cantidad_kg <= 0`, y el reporte la daba 100 % cubierta, líneas y
ramas. Pero cuando cambié `<= 0` por `< 0` (un mutante a propósito), ningún test se puso en rojo.
La cobertura sólo dice que la línea se ejecutó, no que algún test dependa de lo que hace: el 0 ya
lo rechazaba `!item.cantidad_kg`, así que el `<= 0` nunca decidía nada. Un número alto puede
convivir con código que ningún test verifica de verdad. Cómo lo resolví está en *Problemas
encontrados*.

### El Pull Request bloqueado

**PR #20, check `build-frontend`.** Agregué `descuentoMayorista` en
`frontend/src/lib/descuento.js` sin tests: el build pasó, los 21 tests pasaron, y el paso de
cobertura frenó en **las dos métricas**: **79,16 % de líneas** (19 de 24) y **63,63 % de ramas**
(14 de 22), contra el umbral de 90 / 90.

En vitest 5, una función que ningún test llama suma igual sus ramas sin cubrir; por eso cayeron las
dos métricas. En vitest 3, el del video, sólo caían las líneas.

Para arreglarlo escribí `descuento.test.js` con dos `it.each`: uno con los escalones y los bordes
exactos (1, 10, 10,01, 25, 25,01, 50, 50,01 y 200 kg) y otro con las cantidades inválidas (0, -5 y
`NaN`). Volvió a 100 / 100, el check quedó verde y lo mergeé.

El **PR #21** muestra el freno del backend: `Cotizar` en `backend/internal/envio/envio.go`, sin
tests, baja el total a **80,6 % (158 de 196 sentencias)** contra el 85 %, y `build-backend` falla
con `la cobertura de sentencias (80.6%) no llega al umbral (85%)`. Queda abierto y en rojo.

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
- **Sentencias contadas dos veces con `-coverpkg`.** Cada paquete de tests repite los mismos bloques
  en el perfil, y sumando a mano me daba «163 de 342». El script cuenta cada bloque una sola vez: 158 de 168.
- **El script y los finales de línea de Windows (CRLF).** En Windows el `.sh` se guardaría con CRLF y
  el `sh` del contenedor Linux no lo correría. Lo fijé con `.gitattributes` (`*.sh text eol=lf`).
- **vitest 5.** Tengo Vite 8, y vitest 5.0.2 es la versión que lo soporta. `@vitest/coverage-v8`
  tiene que ser exactamente la misma versión (5.0.2). Y desde vitest 4 el `include` de la
  cobertura es obligatorio (ver arriba).
- **Windows Defender bloqueando el repo.** Git no podía crear la rama y los tests de vitest fallaban
  al escribir archivos: era el *Acceso controlado a carpetas*. Permití git, bash y node sin apagar la protección.
- **Un commit de prueba que volvió en el PR #21.** El PR traía, además de `envio.go`, dos líneas
  en blanco en el `README.md`: el commit `2a432ea` («prueba»). En la rama del PR #20 lo había
  revertido, y como ese PR se mergeó con squash, a `main` no llegó nada. Pero mi `main` local
  todavía lo tenía, y al hacer `git pull` git armó un merge que lo trajo de vuelta; la rama del
  PR #21 salió de ahí. Lo revertí en esa rama y dejé mi `main` igual al de GitHub con
  `git reset --hard origin/main`, después de confirmar que no tenía ningún otro commit local. Para
  prevenirlo, antes de crear una rama corro `git log origin/main..main`: si muestra algo, mi `main`
  no es el de GitHub.

### Uso de IA

Usé Claude Code para escribir los tests, los refactors, el script de cobertura y los cambios del
pipeline, y Claude para consultar dudas puntuales. Revisé cada cambio contra la guía antes de
commitearlo. Las decisiones fueron mías: los umbrales y su justificación, qué excluir de la cuenta
(y dejar los handlers adentro), descartar tests duplicados, simplificar la condición de cantidad y
redefinir las zonas de envío desde Catamarca.

## TP6 — CD: environments, aprobaciones y deployment patterns

> 🔗 **Enlaces de este TP**
> - Paquete del backend: https://github.com/users/Magdalenagomezf/packages/container/package/ingsoft3-tp01-backend
> - Paquete del frontend: https://github.com/users/Magdalenagomezf/packages/container/package/ingsoft3-tp01-frontend
> - Corrida de un PR con los tests en verde, donde «Entrar al registry» sale salteado (PR #24, job `build-backend`): https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/38733683973/job/116417469884
> - Lista de pasos de una corrida de `main`, con «Construir y publicar la imagen» como último paso (job `build-backend`): [https://github.com/Magdalenagomezf/ingsoft3-tp01/actions/runs/38588463622/job/115936683810]
> - QA: https://mallki-front-qa.onrender.com (api: https://mallki-api-qa.onrender.com)
> - PROD: https://mallki-front-prod.onrender.com (api: https://mallki-api-prod.onrender.com)

### Por qué la imagen se publica sólo con la verificación en verde

El paso que publica es el último del mismo job que corre los tests. No tiene ningún `if` que diga
«si los tests pasaron»: si los tests fallan, el job se corta ahí y nunca llega a publicar. En el
TP4 el build de la imagen estaba arriba, antes de los tests; lo moví al final. Si lo hubiera dejado
donde estaba, todo iba a seguir en verde y habría estado publicando imágenes que todavía nadie
había verificado, sin ningún error que me avisara.

Son tres eslabones encadenados: nada entra a `main` sin los checks del TP4/TP5 en verde, sólo
`main` publica (`push:` con `github.event_name == 'push' && github.ref == 'refs/heads/main'`), y
publicar es lo último. Si se publicara igual, el registry dejaría de significar «esto pasó la
verificación» y pasaría a ser un depósito de cualquier cosa que alguna vez se construyó.

Lo que **no** garantiza: nada impide que alguien con permisos haga un `docker push` a mano con un
tag `sha-…` que parezca legítimo. Además, el tag se puede volver a apuntar a otra imagen; lo que no
cambia nunca es el *digest* (`sha256:…`), que identifica el contenido exacto.

El nombre de la imagen lo escribí a mano en minúsculas porque mi usuario tiene mayúscula y
`${{ github.repository }}` da `repository name must be lowercase`.

### Continuous Delivery, no Deployment

Implementé **Continuous Delivery**: cada merge llega solo a QA, pero a producción sólo pasa si yo
lo apruebo. Para Mallki tiene sentido porque es un negocio real y los clientes hacen pedidos
reales: prefiero elegir cuándo cambia producción (por ejemplo, no en medio de la toma de pedidos)
y mirar QA antes. Además, mis tests son unitarios: no hay ningún test que recorra la app completa,
así que un deploy automático a PROD dependería de algo que hoy no tengo.

Para Continuous Deployment me faltarían tests end-to-end, un smoke que compruebe qué versión corre
y monitoreo. Por ahora no lo querría.

### Proveedor y el contrato de los cinco puntos

Usé **Render** para los contenedores y **Neon** para Postgres. Los cinco puntos: dos entornos con
su URL (`mallki-*-qa` y `mallki-*-prod`); una base por entorno (`app_qa` y `app_prod`, cada api
recibe la suya en `DATABASE_URL`); los cuatro servicios son `Docker` con mis Dockerfiles; el deploy
lo dispara el pipeline con el commit verificado, con Auto-Deploy en Off en los cuatro; y
producción está detrás de la aprobación del environment `production`.

Mi backend crea las tablas solo al arrancar (`CREATE TABLE IF NOT EXISTS` en `db.go`), así que no
tuve que correr migraciones: las bases se llenaron la primera vez que cada api arrancó en Render.
Lo comprobé con `select count(*) from productos;` en las dos bases: antes daba `relation does not
exist`, después daba un número.

### Cómo comprobé que cada entorno usa su propia base

Inserté un producto `SOY PROD` **sólo en `app_prod`**, desde el SQL Editor de Neon. Después abrí
`/api/productos` en las dos apis: en PROD aparecía y en QA devolvía `[]`. Las dos cadenas de
conexión difieren en una sola palabra, y si las hubiera cruzado nada se habría puesto en rojo.

### Qué quedó por variable y qué quedó dentro de la imagen

| Por variable (la pone cada entorno) | Dentro de la imagen |
|---|---|
| `DATABASE_URL` (api) | el binario de Go |
| `BACKEND_URL` y `DNS_RESOLVER` (front) | el `dist/` de Vite y la plantilla `default.conf.template` |
| `PORT` (lo agregué para que Render sepa el puerto) | los valores por defecto de compose (`ENV` del front) |

Mi `nginx.conf` del TP2 tenía escrito `http://backend:8080` (el nombre del servicio en compose) y
el DNS de Docker. En Render ese nombre no existe, y la api es otra URL distinta en QA y en PROD. Lo
pasé a una plantilla que nginx completa al arrancar con las variables del entorno. Lo probé en mi
máquina con la misma imagen: sin variables quedaban los valores de compose, y con
`-e BACKEND_URL=… -e DNS_RESOLVER=8.8.8.8` quedaban los de QA. Mi `docker compose up` del TP2
sigue andando sin cambios.

### Diseño de la cadena y alcance de los secrets

```
build-backend ─┐
               ├─→ deploy-qa ─→ deploy-prod
build-frontend ┘   (solo)       (espera aprobación)
```

- `deploy-qa`: `needs` de los dos builds (sin CI verde no hay nada que desplegar), `if` de rama
  `main` (los PR verifican pero no despliegan) y `environment: qa`.
- `deploy-prod`: `needs: deploy-qa` y `environment: production`. No repite el `if`: como depende
  de `deploy-qa`, en un PR se saltea igual. Se ve en los checks de cualquier PR: los dos deploys
  salen salteados.
- `concurrency: deploy-prod` para que dos deploys a producción no se pisen. No ordena la cola de
  aprobaciones: si quedan dos corridas esperando, la vieja la tengo que rechazar yo.
- Los cuatro deploy hooks son **secrets de environment**: `RENDER_HOOK_API_QA` y
  `RENDER_HOOK_FRONT_QA` en `qa`, y los de PROD en `production`. En los secrets del repositorio no
  hay ninguno. Así, los hooks de producción sólo los puede leer un job que ya pasó la aprobación.
  Si estuvieran en el repo, cualquier job de cualquier workflow podría desplegar a producción sin
  pasar por mí. Los hooks de PROD tienen el id de su servicio (`srv-…`), y lo comparé con el
  Service ID de cada servicio de producción en Render.
- Los dos `curl` mandan `&ref=$GITHUB_SHA`. Sin eso, el hook despliega la punta de `main`, que
  puede ser un commit que entró mientras corría el pipeline y que nadie verificó. En PROD importa
  más todavía: entre que la corrida queda esperando y que la apruebo pueden pasar horas.

### Qué mira mi aprobador antes de aprobar

1. Que `deploy-qa` esté en verde, y cuántas vueltas tardó el smoke.
2. Que en Render, QA esté **Live con el commit de esa corrida**, y que **sirva** ese build. No
   alcanza con el «Live» (ver *Problemas encontrados*). Lo compruebo con
   `curl.exe -s <url-del-front-de-QA>/ | Select-String "assets/index"` y comparo el nombre del JS
   con el que generó el build en el log de Render.
3. Que el cambio se vea y funcione en QA: navego la pantalla que cambió.

Lo que mi aprobador **no** puede ver: cómo se comporta con tráfico real, si los datos de
producción tienen algo que QA no tiene, ni si el build de PROD va a salir igual al de QA (Render
vuelve a construir). Me falta observabilidad: no tengo métricas ni logs centralizados, y `/health`
no dice qué versión corre.

**Los rechazos.** Rechacé dos corridas:
- La del merge del PR #25 (`3e37a15`): ese merge sólo agregaba el job de PROD al pipeline y no
  traía ningún cambio de la app para los usuarios.
- La del merge del PR #27 (`878f40a`): Render marcaba el deploy de QA como Live, pero QA seguía
  sirviendo el build anterior. El smoke estaba verde y el cambio no estaba en QA.

Después la volví a correr, comprobé que QA servía el build nuevo y la aprobé.

### Límites del plan gratis y cómo los maneja el pipeline

- **Render**: los servicios duermen a los 15 minutos sin tráfico, y el primer pedido los despierta.
  Medí un despertar de unos 26 segundos. Las 750 horas por mes son por workspace, repartidas entre
  mis cuatro servicios. Los builds son lentos: el primero de la api de PROD tardó 10 minutos.
- **Neon**: gratis sin vencimiento, 0,5 GB, se suspende a los 5 minutos y se despierta solo. El
  Postgres gratis de Render vence a los 30 días.
- **Cómo lo maneja el pipeline**: el smoke reintenta hasta 30 veces cada 20 segundos, y cada `curl`
  tiene `--max-time 10` para que un servicio que se está despertando no deje el job colgado. En el
  log del primer deploy a QA se ve: una vuelta con `Operation timed out after 10002 milliseconds`
  y la siguiente con `QA responde`.


### Qué prueba mi smoke test y qué no

Prueba tres cosas: que la api está viva (`/health`), que la api llega a su base
(`/api/productos`, porque un `/health` solo daría verde con la conexión rota) y que el front se
sirve.

No prueba **qué versión** está corriendo. Me pasó en el PR #27: el smoke dio verde y Render marcaba
`878f40a` como Live en `mallki-front-qa`, pero la URL seguía sirviendo el build anterior.
Tampoco prueba que el front llegue a la api a través de nginx, porque al front sólo le pide la
página. Para cerrarlo, `/health` tendría que devolver el commit que corre y el smoke tendría que
compararlo con `github.sha`.

### Deployment pattern para producción real y plan de rollback

**Pattern: blue-green.** Mallki tiene pocos clientes y pocos pedidos por día, así que no hay
tráfico suficiente para que un canary al 5 % me diga algo: serían uno o dos pedidos. Con
blue-green pruebo la versión nueva al lado de la vieja, cambio todo el tráfico de una vez y, si
algo sale mal, vuelvo el switch en segundos.

El costo es tener la producción duplicada: cuatro servicios en lugar de dos, siempre prendidos
(en el plan gratis no se puede, porque se duermen). Hoy, con pocos pedidos por día, una caída
corta no le cuesta mucho a Mallki, así que no lo implementaría todavía. Lo elegiría cuando el
negocio crezca y una caída en horario de pedidos empiece a costar ventas. Otra cosa a tener en
cuenta: azul y verde comparten la misma base, así que cualquier cambio en las tablas tiene que
funcionar con las dos versiones a la vez.

Para hacerlo me falta un balanceador que haga el cambio de tráfico, y monitoreo (errores y
tiempos de respuesta) para darme cuenta rápido de que la versión nueva anda mal y tengo que
volver.

**Mi rollback de hoy**, medido: el 15/10 volví PROD de `878f40a` a `v6.0.0` (`831c685`).

1. `git rev-list -n1 v6.0.0` para sacar el sha de la última versión buena.
2. Copio de Render los deploy hooks de PROD (el secret de GitHub no se puede volver a leer).
3. Llamo los dos hooks con `&ref=<ese sha>`.
4. Espero en Render → *Deploys* a que ese commit quede Live en los dos servicios.
5. Compruebo que el front sirve el build viejo (`index-DWBZ0Fyk.js` en lugar de
   `index-BAnfCuL3.js`), no sólo que Render diga Live.

En *Deploys* de Render, el deploy de `831c685` (disparado por Deploy Hook) tardó **25,1 s** en la
api y **30,3 s** en el front. PROD volvió entero en unos **30 segundos**, desde las
[10:59] hasta las [11:01]. Fue rápido porque los servicios estaban despiertos; dormidos se suma el 
despertar. Es la métrica DORA *failed deployment recovery time*.

Lo que el rollback **no** deshace son los **datos**: si un deploy hubiera borrado una columna o
modificado registros de `app_prod`, volver el código no los recupera. Para eso haría falta un
backup de la base antes de cada deploy, o migraciones que se puedan revertir. Además, este rollback
no aparece en *Deployments* de GitHub, porque lo disparé a mano y no desde el pipeline.

### Problemas encontrados y cómo los resolví

- **El primer deploy de la api de PROD quedó colgado.** Estuvo más de una hora en *Building* y
  después Render dio `Port scan timeout reached, no open ports detected`, sin ningún log de mi app.
  Mi backend se conecta a la base **antes** de abrir el puerto, así que si la conexión se queda
  esperando, Render nunca ve el puerto abierto. Agregué `&connect_timeout=10` a `DATABASE_URL` (en
  los dos entornos) para que, si la base no responde, la app falle a los 10 segundos con un mensaje
  en vez de colgarse. Con eso arrancó y apareció `servidor escuchando en :8080`. En el medio cancelé
  varios deploys sin querer: cada vez que guardaba una variable, Render arrancaba otro y cancelaba
  el anterior.
- **`channel_binding` en la cadena de Neon.** Neon la da con `&channel_binding=require`, que es
  para drivers de .NET. Lo saqué y dejé sólo `?sslmode=require`, porque `lib/pq` (el driver de
  Postgres que usa mi backend) no lo reconoce.
- **Render decía Live pero servía la versión anterior** (PR #27). Me di cuenta porque el subtítulo
  nuevo no aparecía en QA ni en incógnito. El log del build de Render mostraba que había bajado
  `878f40a` y generado `index-BAnfCuL3.js`, pero el `index.html` de la URL pedía
  `index-DWBZ0Fyk.js`. Con `curl` busqué el texto nuevo en ese JS y no estaba. Rechacé PROD, Restart service en Render, verifiqué que
  QA sirviera `BAnfCuL3` y recién ahí aprobé. Mientras investigaba me confundió algo de mi propio
  nginx: `try_files … /index.html` devuelve `200` para cualquier archivo que no existe, así que un
  `200` no probaba que el JS nuevo estuviera.
### Uso de IA
Usé Claude para traducir la guía y el video (escritos sobre .NET) a mi stack en Go, para guiarme
paso a paso y para diagnosticar los problemas de Render. También me ayudó a redactar este
documento a partir de lo que fuimos haciendo; lo revisé y lo ajusté. Cada paso lo hice yo y lo
comprobé en GitHub, Render y Neon antes de seguir. Las decisiones de qué aprobar y qué rechazar,
y por qué, fueron mías.

## TP7 — Contenedores en el pipeline + integración y e2e

Los enlaces de este TP están arriba de todo, en «Enlaces del TP7».

### Build once, deploy many

En el TP6, Render reconstruía la app desde el repo en cada deploy: QA construía con el merge y PROD
volvía a construir cuando yo aprobaba, a veces horas después. Mi Dockerfile del front arranca con
`FROM nginx:alpine`, y `alpine` es una etiqueta que se mueve: si entre esos dos builds salía una
versión nueva de nginx, PROD iba a correr algo que QA nunca probó, con el mismo código.

Ahora la imagen se construye **una sola vez**, en el pipeline, después de los tests, y se publica
como `sha-<commit>`. QA y PROD **ejecutan esa misma imagen**: Render ya no construye, sólo la baja.

Esto también resuelve algo que anoté en el TP6, en *Qué mira mi aprobador*: no podía saber si el
build de PROD iba a salir igual al de QA. Ya no hay un segundo build.

Para publicar no agregué ningún job: los pasos que publican las imágenes ya estaban desde el TP6.
Lo que cambió es quién las ejecuta.

### Cómo pasé los servicios a imagen

A los cuatro servicios de Render (api y front, en QA y en PROD) no los creé de nuevo: les cambié la
fuente a *Existing Image*. Así conservaron su URL, sus variables y su deploy hook, y no tuve que
volver a cargar ningún secret ni cambiar las URLs del smoke.

Antes de tocar Render comprobé que las dos imágenes fueran públicas: hice `docker logout ghcr.io`
y después `docker pull` de cada una, y las dos bajaron.

Lo hice en orden: primero sólo `mallki-front-qa`, y lo desplegué una vez a mano con su hook y el
`imgURL`. En *Events* apareció «Triggered via Deploy Hook» con el `sha-`, sin ningún build. Después
les cambié la fuente a los otros tres, y recién al final cambié el pipeline (PR #29).

En el `ci.yml`, el hook dejó de mandar `&ref=$GITHUB_SHA` («construí este commit») y pasó a mandar
`imgURL=ghcr.io/magdalenagomezf/ingsoft3-tp01-<backend|frontend>:sha-${{ github.sha }}` («ejecutá
esta imagen»). QA y PROD reciben exactamente la misma imagen.

### Estrategia de etiquetas

- **`sha-<commit>`, con los 40 caracteres, en el registry.** Nombra la imagen de ese commit. La
  publica sólo mi pipeline, con el `GITHUB_TOKEN` de la corrida, pero no es inmutable: volver a
  correr la corrida de un commit la republica, y con un `docker push` a mano alguien con permisos
  podría pegarla en otra imagen; mi cadena no se daría cuenta y desplegaría algo que nadie verificó.
  Por eso no re-corro corridas viejas para desplegar. Lo que no se mueve es el *digest*
  (`sha256:…`), y el `imgURL` del hook lo acepta en lugar de la etiqueta. Tiene que ir completa:
  git completa un commit corto, pero el registry compara el texto letra por letra.
- **No publico `latest`.** Es un rótulo que se mueve con cada publicación y no dice qué hay detrás:
  un deploy por `latest` ejecutaría «lo último que se publicó», haya pasado las pruebas o no.
- **`v7.0.0` en git.** Es el nombre de una versión que sé que anduvo en PROD.
- **Con qué imagen configuré Render, y por qué no es la que corre.** Para cambiar la fuente, Render
  pide una imagen que exista. Le puse la de mi último merge de ese momento
  (`sha-705fad4f1e00ab5395a91a731cd8022441fc95a1`). Es sólo el punto de partida: lo que corre lo
  decide cada corrida, con el `imgURL` del hook. Por eso no uso *Manual Deploy* desde el panel:
  despliega la imagen que quedó en *Settings*, que ya es vieja, y llevaría el entorno para atrás
  sin que nada se ponga en rojo.

### De la release a la imagen

Puse el tag `v7.0.0` sobre el commit que me dio *Deployments* para `production`
(`gh api 'repos/{owner}/{repo}/deployments?environment=production' --jq '.[0].sha'`), no sobre el
último de `main`. Esta vez coincidían, pero no tiene por qué: si después de aprobar entra otro
merge, el último de `main` nunca llegó a PROD.

Del tag a la imagen hay un paso: `git rev-list -n1 v7.0.0` da
`fc3fdfeb7059c0a1d454f809257ad096afbf54ac`, y los dos paquetes tienen la imagen
`sha-fc3fdfeb7059c0a1d454f809257ad096afbf54ac`.

Volver a esta versión es mandarle esa imagen al hook. En el TP6, para el rollback había que
reconstruir desde el commit viejo. Ahora Render no construye nada: baja una imagen que ya existe.

### Cómo compruebo que un entorno ejecuta mi imagen

La respuesta del hook (`{"deploy":{"id":"dep-..."}}`) sólo dice que Render **aceptó** el pedido, no
que la imagen ya esté corriendo. El smoke tampoco alcanza: pregunta si el entorno responde, y si el
deploy hubiera fallado, Render sigue sirviendo la versión anterior, que también responde.

Lo compruebo en *Events* de cada servicio: el deploy tiene que decir «Triggered via Deploy Hook» y
nombrar el `sha-` de esa corrida, y no tiene que haber ningún build. Cerrarlo automáticamente (que
`/health` devuelva el commit y el smoke lo compare, como anoté en el TP6) sigue pendiente.

### Una función nueva para que las pruebas puedan limpiar

Las pruebas que crean datos tienen que borrarlos, porque QA es uno solo y lo comparten todas las
corridas. En mi app eso no se podía: no había forma de borrar un pedido, ni desde la api ni desde
la pantalla, y cada pedido descuenta stock. Si la e2e armaba un pedido en cada corrida, QA se iba a
llenar de pedidos y el stock iba a bajar hasta cero.

Agregué **Cancelar pedido** (PR #30): `DELETE /api/pedidos/{id}` borra el pedido y sus items y
devuelve el stock, todo en una transacción, y en la pantalla Pedidos hay un botón «Cancelar» en
cada pedido. Además le sirve al negocio: hoy un pedido cargado por error no se podía deshacer.

Además, cada prueba que necesita un producto lo crea con un nombre que no se repite
(`e2e producto <hora>`) y un stock conocido, y lo borra al final. Así no dependen de los productos
que haya en QA.

### Qué probé en cada suite, y por qué ésas

**Integración** (`frontend/e2e/api.spec.js`, sin navegador, contra la api de QA):

1. Alta de un producto → el `GET` lo encuentra → `DELETE` → el `GET` ya no lo encuentra.
2. Producto con nombre vacío → `400`, y la lista no cambió. El precio y el stock van válidos a
   propósito, para que el único motivo del rechazo sea el nombre.
3. **La que elegí:** pedido con más kg que el stock → `400` con «stock insuficiente», el stock
   sigue igual y no apareció ningún pedido. Es la regla más cara del negocio: si se rompe, vendemos
   nueces que no tenemos, y el que se entera es mi papá. Además prueba que la transacción hace
   rollback en Postgres de verdad, algo que mis unitarios con `sqlmock` no pueden probar.

**E2E** (`frontend/e2e/pedidos.spec.js`, Chromium contra el front de QA):

1. **Creación:** armar un pedido → verlo en Pedidos → cancelarlo → ya no está, y el stock volvió.
2. **Error:** pedir 5 kg de un producto que tiene 1 kg → el usuario ve el error de stock → no se
 creó el pedido.
3. **La que elegí:** un pedido confirmado sigue en Pedidos al recargar, y cancelado no vuelve. Es
  lo que hace el cliente todos los días: si el pedido no queda guardado, se pierde la venta.

Para que Playwright encontrara los controles como lo haría una persona (`getByLabel`, `getByRole`),
tuve que agregarle a la app `aria-label` en el input de kg y en el botón «Agregar» de cada producto
(antes eran todos iguales), y `role="alert"` en los mensajes de error.

**Qué no puse en cada suite.** En e2e no repito lo que ya cubren los unitarios del TP5 (nombre o
contacto vacíos, cantidades en cero, el total del carrito, los descuentos) ni cada error posible de
la api: abajo es más barato y más rápido. En integración no pruebo pantallas ni el carrito, que son
del front. Y la e2e no prueba todos los caminos: son tres flujos, los que si se rompen la app no
sirve.

### Lo que ve la integración y no ve el unitario

**Un cambio de una línea que sólo atrapa la integración.** Si en `producto/repository.go` escribo
mal una columna del `INSERT` (por ejemplo `stok_kg` en vez de `stock_kg`), los unitarios siguen en
verde: los tests de producto usan un `mockRepo` y nunca ejecutan ese SQL. Postgres, en cambio,
contesta `column "stok_kg" does not exist`, la api da `500`, y mi prueba de integración 1 se pone en
rojo.

Es lo que anoté en el TP5: excluí `repository.go` de la cobertura porque «su lugar natural es un
test de integración contra una base real». Ahora lo tiene.

### Integración y e2e no prueban lo mismo: el par verde/rojo

La integración le habla a la api directamente, con su base real. La e2e usa la app por la pantalla,
así que además prueba cómo el front usa la api. Por eso corren en ese orden
(`deploy-qa → integracion → e2e → deploy-prod`): si la api está rota, la integración lo ve en
segundos y la e2e ni corre.

Para la evidencia rompí **el front** a propósito (PR #33, el commit `681cb30`): en `api.js`,
`createPedido` pasó a llamar a `/pedido` en vez de `/pedidos`. No rompí el nombre de un campo,
porque mis unitarios del TP5 comprueban qué manda `confirmarPedido` y lo habrían atajado antes de
llegar a QA.

El resultado fue: unitarios en verde, smoke en verde, **integración en verde (11 s)**, **e2e en
rojo (2 min 17 s)** y `deploy-prod` salteado, sin pedirme aprobación.

Lo diagnostiqué con los dos reportes, sin abrir el código. En la traza de la e2e (pestaña
*Network*), al tocar «Confirmar pedido» el navegador mandó `POST /api/pedido` y la api contestó
`404`; el test esperó 15 s la confirmación y no apareció. En el reporte de integración, la api con
la ruta correcta funcionaba. **La api y la base estaban sanas; el que no usaba bien la api era el
front** (fila 1 de la tabla de §2.5: integración verde y e2e roja apuntan al front).

Si hubiera roto la api, la integración se habría puesto en rojo y la e2e no habría corrido, porque
depende de ella (fila 2).

Lo arreglé en el PR #34, la cadena quedó en verde hasta PROD, y sobre ese commit puse `v7.0.0`.

Lo que este gate **no** ataja: lo que mis tres flujos no recorren, y lo que sólo pasa en PROD (otra
base, otros datos, más carga).

### Qué mira mi aprobador ahora

En el TP6 aprobaba sabiendo sólo que QA respondía. Ahora, cuando `deploy-prod` me pide aprobación,
ya pasó que la api contestó bien con su base y que un navegador usó la app entera sobre la misma
imagen que va a PROD. Igual reviso en *Events* que QA esté corriendo el `sha-` de esa corrida antes
de aprobar.

En el TP6 escribí que para Continuous Deployment me faltaban tests end-to-end. Ahora los tengo,
pero sigo con la aprobación manual: me sigue faltando que el smoke compruebe qué versión corre, y
monitoreo en PROD.

### Por qué integración amplia y no estrecha

Mi integración corre contra la api **ya desplegada en QA**, que ejecuta la imagen de esa misma
corrida. La otra forma era la estrecha: levantar la api y un Postgres descartable dentro del job.

Con la amplia gano mucha menos configuración, y pruebo la imagen real, con la red y el proveedor en
el medio. Pierdo velocidad, porque tiene que esperar el deploy, y aislamiento: QA es uno solo, y
dos corridas se pueden pisar.

### Cold start, timeouts y tests flaky

Render gratis duerme los servicios. Lo vi el primer día: después de desplegar el front de QA a
mano, los productos daban `502` porque la api de QA estaba dormida; al despertarla, anduvo.

Por eso la config de Playwright tiene 60 s por test, 15 s por aserción y un reintento. Las esperas
son de Playwright (`toBeVisible`, `toHaveCount`), sin `sleep`. Además, el smoke de `deploy-qa`
corre antes y despierta QA.

Un test **flaky** pasa o falla sin que cambie nada. Es peor que no tenerlo, porque el equipo se
acostumbra a ignorar el rojo. Si un test pasa recién en el reintento, Playwright lo marca como
flaky y la corrida queda en verde: por eso abro el reporte igual.

### La misma imagen del front en QA y en PROD

La imagen del front no sabe a qué api llamar: nginx manda `/api` a `BACKEND_URL`, que es una
variable de cada servicio en Render, junto con `DNS_RESOLVER` (lo armé así en el TP6). QA y PROD
corren la misma imagen, y lo único que cambia es esa variable. Al pasar los servicios a imagen no
toqué ninguna variable.

### Límite conocido: QA compartido

QA es uno solo, y ahora lo usan dos jobs seguidos. Si mergeo dos veces seguidas, la corrida B puede
redesplegar QA mientras las pruebas de A están corriendo, y A probaría en parte la imagen de B. Lo
reconocería mirando en *Events* qué imagen corre y a qué hora fue el `deploy-qa` de B. En ese caso
rechazaría la A con el motivo, y vale la B, que trae los dos cambios. Por eso hago un merge por vez
y espero a que termine la cadena antes del siguiente.

### Problemas encontrados y cómo los resolví

- **`Invalid URL` en todas las e2e locales.** Había creado `playwright.config.js` fuera de
  `frontend/`, así que Playwright corría sin config: sin `baseURL` y sin reintentos (me di cuenta
  porque no aparecía ningún `retry #1`). Lo moví a `frontend/`.
- **Una e2e que falla a la mitad no llega a limpiar.** La limpieza está al final de cada test, así
  que si falla antes, no se ejecuta: después de la corrida roja quedaron 5 productos
  `e2e producto …` en QA. Los borré por la api filtrando por ese nombre. Es un límite conocido: lo
  detecto listando `/api/productos` de QA y buscando ese prefijo. Para resolverlo de verdad habría
  que mover la limpieza a un `afterEach`, y en la e2e cancelar primero el pedido (si se creó),
  porque un producto con pedidos no se puede borrar.
- **`gh run download` no pudo descomprimir los reportes en Windows.** Los bajé desde *Artifacts*
  en la web.

### Uso de IA

Usé Claude para adaptar la guía y los videos a mi stack y para guiarme paso a paso; revisamos cada
diff antes de commitear. Claude Code implementó «Cancelar pedido» (backend, front y tests
unitarios) a partir de un pedido mío; revisé el diff y lo probé a mano, local y en QA, antes de
mergear.

Los specs de Playwright y los jobs del pipeline los escribió Claude partiendo de los de la cátedra
y de mi app; los corrí yo, local y en el pipeline, y puedo explicar qué verifica cada uno, contra
qué entorno corre y qué pasa si falla. Este documento también lo redactó Claude a partir de lo que
hicimos; lo revisé y lo ajusté. Qué aprobar y qué rechazar lo decidí yo.

## TP8 — Infraestructura como código

### Riel y versiones

- Riel canónico: **Terraform + Neon + Render**. Terraform v1.16.6 en Windows (instalado con winget).
- Providers fijados por versión: `kislerdm/neon` **0.18.0** y `render-oss/render` **1.9.1**, con el
  `.terraform.lock.hcl` commiteado. El estado vive en mi máquina y está en el `.gitignore`.
- Preprod es un entorno nuevo: proyecto de Neon `mallki-preprod` (rol + base `app_preprod`) y
  `mallki-api-preprod` / `mallki-front-preprod` en Render (plan free, región `virginia`, como QA y
  PROD), corriendo la imagen `sha-5aa413c…` del TP7. QA y PROD no se tocaron.
- Adaptado a mi app: mi backend en Go lee la base de `DATABASE_URL` en formato
  `postgres://usuario:clave@host/base?sslmode=require` (no `ConnectionStrings__Default`), y el
  endpoint que toca la base es `/api/productos`. La cadena se arma en un `locals` con lo que
  devuelve Neon, y la `BACKEND_URL` del front sale del `url` de su propia api.

### Problemas encontrados y cómo los resolví

- **winget instaló Terraform pero no creó el alias**, así que `terraform` no se reconocía ni
  reiniciando la terminal. Agregué la carpeta del paquete al PATH de usuario.
- **`terraform init` fallaba con `mkdir .terraform: El sistema no puede encontrar el archivo
  especificado`.** Ni PowerShell podía crear carpetas ahí: era el *Acceso controlado a carpetas* de
  Windows. Agregué `terraform.exe` y `powershell.exe` como aplicaciones permitidas.
- **El primer `apply` falló con `org_id is required`**: mi cuenta de Neon trabaja con una
  organización. Agregué `org_id` al `neon_project` (no es un secreto).
- **En el `.gitignore` puse el comentario de `*.tfvars` en su propia línea.** En `.gitignore` el
  `#` solo comenta al principio de la línea; escrito al final, el patrón deja de ignorar los tfvars.
- **Medido al recrear preprod (destroy → apply):** Render devolvió las mismas URLs, sin sufijo. El
  host de Neon cambió, y como la cadena de conexión se deriva, no hubo que tocar nada.

### Uso de IA

Usé Claude para seguir la guía y los videos paso a paso, y para adaptarlos a mi app (variable de
conexión de Go, región, nombres). El HCL sale de los snippets de la guía con esas adaptaciones.
Corrí yo cada comando, leí cada plan antes de aplicar y verifiqué el resultado en las consolas de
Neon y Render.