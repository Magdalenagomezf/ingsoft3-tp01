#!/bin/sh
# Corre la suite del backend con cobertura y aplica el umbral.
# Es la misma receta en la máquina local y en el pipeline:
#   sh scripts/coverage.sh              -> reporte en ./coverage-out
#   sh scripts/coverage.sh /out         -> reporte en /out (etapa test del Dockerfile)
# Sale con error si fallan los tests, si no se midió ningún archivo
# o si el total de sentencias cubiertas queda por debajo de UMBRAL.
set -eu

UMBRAL=85
OUT="${1:-./coverage-out}"

# Se excluye por archivo (lo que se saca, no lo que entra): un archivo nuevo
# entra solo a la cuenta. Ver decisiones.md, TP5.
EXCLUIR='/main\.go:|/db\.go:|/model\.go:|/repository\.go:'

mkdir -p "$OUT"
RAW="$OUT/coverage.raw.out"
PERFIL="$OUT/coverage.out"
RESUMEN="$OUT/resumen.md"

# -coverpkg=./... cuenta también el código que se ejecuta desde otro paquete
# (httpx lo usan los handlers). Sin eso daría 0 % aunque se ejecute.
set +e
go test ./... -coverpkg=./... -coverprofile="$RAW"
TESTS=$?
set -e

if [ ! -s "$RAW" ]; then
  echo "ERROR: go test no generó el perfil de cobertura" >&2
  printf '### Coverage del backend\n\n**Sin perfil de cobertura**: fallaron los tests antes de medir.\n' > "$RESUMEN"
  exit 1
fi

# Primera línea (mode: ...) + los bloques de los archivos que no se excluyen.
head -n 1 "$RAW" > "$PERFIL"
tail -n +2 "$RAW" | grep -vE "$EXCLUIR" >> "$PERFIL" || true

BLOQUES=$(tail -n +2 "$PERFIL" | wc -l | tr -d ' ')
if [ "$BLOQUES" -eq 0 ]; then
  echo "ERROR: midió 0 archivos; revisá el filtro EXCLUIR" >&2
  printf '### Coverage del backend\n\n**Midió 0 archivos**: revisá el filtro de exclusiones.\n' > "$RESUMEN"
  exit 1
fi

go tool cover -html="$PERFIL" -o "$OUT/coverage.html"
go tool cover -func="$PERFIL" > "$OUT/coverage.func.txt"

TOTAL=$(awk '/^total:/ { sub("%", "", $NF); print $NF }' "$OUT/coverage.func.txt")
# Con -coverpkg cada bloque aparece una vez por paquete de tests: se cuenta una
# sola vez, y queda cubierto si algún paquete lo ejecutó.
SENTENCIAS=$(tail -n +2 "$PERFIL" | awk '{ n[$1] = $2; if ($3 > 0) c[$1] = 1 } END { for (k in n) { t += n[k]; if (k in c) s += n[k] } printf "%d de %d", s, t }')
ARCHIVOS=$(tail -n +2 "$PERFIL" | cut -d: -f1 | sort -u | wc -l | tr -d ' ')

if awk -v t="$TOTAL" -v u="$UMBRAL" 'BEGIN { exit !(t + 0 >= u + 0) }'; then
  ESTADO="pasa"
else
  ESTADO="**NO pasa**"
fi
if [ "$TESTS" -ne 0 ]; then
  ESTADO="**tests en rojo**"
fi

{
  echo "### Coverage del backend"
  echo ""
  echo "| métrica | valor |"
  echo "|---|---|"
  echo "| sentencias | ${TOTAL}% (${SENTENCIAS}) |"
  echo "| archivos medidos | ${ARCHIVOS} |"
  echo "| umbral | ${UMBRAL}% |"
  echo "| resultado | ${ESTADO} |"
  echo ""
  echo "Excluidos: \`main.go\`, \`db.go\`, \`model.go\`, \`repository.go\`. Go mide sentencias, no ramas."
} > "$RESUMEN"

cat "$RESUMEN"

if [ "$TESTS" -ne 0 ]; then
  echo "ERROR: fallaron los tests" >&2
  exit "$TESTS"
fi

if [ "$ESTADO" != "pasa" ]; then
  echo "ERROR: la cobertura de sentencias (${TOTAL}%) no llega al umbral (${UMBRAL}%)" >&2
  exit 1
fi
