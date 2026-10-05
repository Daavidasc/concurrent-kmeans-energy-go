# Informe de análisis con IA: concurrent-kmeans-energy-go

- **Modelo de IA:** Claude Sonnet 5.5 (Anthropic)
- **Fecha:** 4 de octubre de 2026
- **Repositorio / commit analizado:** github.com/Daavidasc/concurrent-kmeans-energy-go, `main`, commit `5c69f52`
- **Entorno de verificación:** Linux, 1 núcleo, Go 1.22 (go.mod ajustado en una copia local), Spin 6.5.2

## Resumen ejecutivo
El proyecto compila, pasa `go vet` y no presenta data races bajo `go test -race`. El diseño del Worker Pool (buffers locales por bloque y fusión ordenada por ID) es sólido. Las brechas principales son de robustez ante entradas límite (K mayor que las filas, 0 filas, NaN), ausencia de pruebas automatizadas y un modelo Promela cuyas aserciones no podían fallar.

## Metodología
Lectura de todo el código, compilación, `go vet`, `gofmt -l`, pruebas temporales con `-race` (20,000 × 48, workers 1, 2, 3, 4, 8 y 16), pruebas de casos borde con tiempo límite y verificación del Promela con Spin (modelo original y modelo extendido con LTL).

## GAPs
| ID | Categoría | Sev. | Dónde | Evidencia | Recomendación |
|---|---|---|---|---|---|
| GAP-01 | Robustez | Alta | `InicializarCentroides` | Con K > filas (K=4, 3 filas) el bucle no termina; con 0 filas, panic `invalid argument to Intn` | Validar `0 < k <= Filas` y devolver error |
| GAP-02 | Robustez | Alta | `procesarAsignacionConcurrente` | Con 0 filas, panic `integer divide by zero` (`tamanoBloque = 0`) | Validar `Filas > 0` antes de crear el pool |
| GAP-03 | Calidad de datos | Media | `CargarDataset` | `ParseFloat` acepta NaN/Inf; con un NaN la inercia es NaN y no hay error | Rechazar valores no finitos indicando fila y columna |
| GAP-04 | Pruebas | Media | repo | 0 archivos `*_test.go`; la equivalencia solo se comprueba en el benchmark manual | Pruebas de equivalencia (varios W) y casos borde; CI con `-race` |
| GAP-05 | Metodología | Media | `main.go`, `benchmark.go` | Secuencial primero y concurrentes después, sin intercalar: deriva térmica/frecuencia (1.ª ejecución secuencial 70.87 s vs ~53.5 s típico) | Intercalar configuraciones, descartar calentamiento, reportar mediana |
| GAP-06 | Concurrencia | Media | `concurrent.go` | Un panic en un worker termina todo el proceso; no hay `context` ni cancelación | `recover` por worker o `context.Context`; propagar errores |
| GAP-07 | Concurrencia | Baja | `concurrent.go` | El pool se crea y destruye en cada una de las 67 iteraciones (costo no medido) | Pool persistente, tras medir el overhead |
| GAP-08 | Configuración | Baja | `main.go` | K, semilla, repeticiones y ruta `./notebooks/data/kmeans_input.csv` son constantes; solo corre desde la raíz | Paquete `flag` |
| GAP-09 | Memoria | Baja | `data.go` | `make([]float32, 0, 3454475*48)` reserva ≈ 663 MB de capacidad aunque el CSV sea pequeño | Capacidad dinámica o conteo previo de filas |
| GAP-10 | Estilo | Baja | `src/` | `gofmt -l` lista los 6 archivos (en `common.go`, falta el salto de línea final) | `gofmt -w` y verificación en CI |
| GAP-11 | Promela | Alta | `promela/kmeans_sync.pml` | `atomic` hace invisible a `updating`: los `assert(!updating)` no pueden fallar; no hay LTL | Modelo con `busy`, `resultsIn`, `ended` y 3 fórmulas LTL (`kmeans_sync_tp.pml`) |
| GAP-12 | Git | Media | repo | `git shortlog -sne` muestra 6 nombres para 3 personas | `.mailmap` y `user.name` único |
| GAP-13 | Repositorio | Baja | `README`, `docs/` | README con PC2 sin marcar y dataset desactualizado; PDF de participación duplicado; faltan los informes; rama `kmena-brand` obsoleta | Actualizar README, limpiar `docs/` y borrar la rama |

## Hallazgos positivos
- `go build` y `go vet ./...` sin advertencias.
- `go test -race` sin `DATA RACE`; iteraciones, inercia (≤ 1e-4) y conteos idénticos al secuencial para W = 1, 2, 3, 4, 8 y 16.
- Sin memoria compartida escrita: buffers locales por bloque y fusión ordenada por ID, de modo que el resultado es reproducible para un W dado.
- Solo librería estándar (0 dependencias).
- `CargarDataset` valida cabecera, columnas por fila y errores de parseo, con el número de fila.
- Con más workers que bloques (5 filas, 64 workers) no falla.
- Modelo Promela del TP: 0 errores en exclusión mutua, barrera y terminación; los 3 mutantes plantados sí son detectados.
- Seguridad: sin red, sin `exec`, sin `unsafe`, sin secretos y sin dependencias. Riesgo residual bajo: el CSV se lee sin tope de tamaño (no probado).

## Recomendaciones priorizadas
1. Corregir GAP-01, 02 y 03 con sus pruebas.
2. Añadir pruebas de equivalencia y CI con `go test -race`.
3. Intercalar y repetir el benchmark.
4. Higiene de Git (`.mailmap`, README, `docs/`).
5. Resto (flags, pool persistente, `gofmt`).

## Límites del análisis
No se ejecutó el benchmark completo (3.45M filas). El entorno tenía 1 núcleo, por lo que no mide speedup. Se usó Go 1.22, no la versión de `go.mod`. El modelo Promela es acotado y abstrae el cálculo. La IA puede equivocarse: los hallazgos se verificaron con comandos y deben reproducirse en el equipo del grupo.