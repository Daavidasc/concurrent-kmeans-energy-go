package kmeans

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"time"
)

// ResultadoBenchmark almacena las mediciones realizadas
// para una determinada configuración.
type ResultadoBenchmark struct {
	Nombre         string
	Workers        int
	Tiempos        []float64
	Media          float64
	MediaRecortada float64
	Speedup        float64
	Eficiencia     float64
}

// calcularMedia calcula el promedio de los tiempos registrados.
func calcularMedia(tiempos []float64) float64 {
	var suma float64

	for _, tiempo := range tiempos {
		suma += tiempo
	}

	return suma / float64(len(tiempos))
}

// calcularMediaRecortada elimina el tiempo mínimo
// y el tiempo máximo antes de calcular el promedio.
//
// Con 10 ejecuciones se elimina:
// 10% inferior + 10% superior.
func calcularMediaRecortada(tiempos []float64) float64 {

	if len(tiempos) <= 2 {
		return calcularMedia(tiempos)
	}

	copia := make([]float64, len(tiempos))
	copy(copia, tiempos)

	sort.Float64s(copia)

	// Eliminar mínimo y máximo.
	recortados := copia[1 : len(copia)-1]

	return calcularMedia(recortados)
}

// resultadosEquivalentes verifica que la ejecución
// concurrente produzca el mismo resultado que la secuencial.
func resultadosEquivalentes(
	referencia ResultadoKMeans,
	actual ResultadoKMeans,
) bool {

	// Deben converger en la misma cantidad de iteraciones.
	if referencia.Iteraciones != actual.Iteraciones {
		return false
	}

	// Permitimos una diferencia mínima debido
	// a operaciones de punto flotante.
	diferenciaInercia :=
		math.Abs(
			referencia.Inercia -
				actual.Inercia,
		)

	if diferenciaInercia > 0.0001 {
		return false
	}

	if len(referencia.Conteos) != len(actual.Conteos) {
		return false
	}

	// Verificar distribución de clusters.
	for i := 0; i < len(referencia.Conteos); i++ {

		if referencia.Conteos[i] != actual.Conteos[i] {
			return false
		}
	}

	return true
}

// BenchmarkSecuencial ejecuta múltiples veces
// el K-Means secuencial.
//
// Devuelve:
//
//  1. Las estadísticas del benchmark.
//  2. Un resultado de referencia para validar
//     las ejecuciones concurrentes.
func BenchmarkSecuencial(
	dataset *Dataset,
	centroidesIniciales []float64,
	k int,
	maxIteraciones int,
	tolerancia float64,
	repeticiones int,
) (ResultadoBenchmark, ResultadoKMeans) {

	tiempos :=
		make([]float64, 0, repeticiones)

	var resultadoReferencia ResultadoKMeans

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(" BENCHMARK SECUENCIAL")
	fmt.Println("========================================")

	for ejecucion := 1; ejecucion <= repeticiones; ejecucion++ {

		// Ejecutar GC antes de medir para reducir
		// interferencias entre ejecuciones.
		runtime.GC()

		inicio := time.Now()

		resultado :=
			KMeansSecuencial(
				dataset,
				centroidesIniciales,
				k,
				maxIteraciones,
				tolerancia,
			)

		duracion :=
			time.Since(inicio).Seconds()

		tiempos =
			append(
				tiempos,
				duracion,
			)

		// La primera ejecución será nuestra referencia.
		if ejecucion == 1 {
			resultadoReferencia = resultado
		}

		fmt.Printf(
			"Ejecución %d/%d: %.4f segundos\n",
			ejecucion,
			repeticiones,
			duracion,
		)
	}

	media :=
		calcularMedia(tiempos)

	mediaRecortada :=
		calcularMediaRecortada(tiempos)

	resultadoBenchmark :=
		ResultadoBenchmark{
			Nombre:         "Secuencial",
			Workers:        1,
			Tiempos:        tiempos,
			Media:          media,
			MediaRecortada: mediaRecortada,
			Speedup:        1.0,
			Eficiencia:     1.0,
		}

	return resultadoBenchmark, resultadoReferencia
}

// BenchmarkConcurrente ejecuta múltiples veces
// K-Means utilizando una cantidad determinada de workers.
func BenchmarkConcurrente(
	dataset *Dataset,
	centroidesIniciales []float64,
	k int,
	maxIteraciones int,
	tolerancia float64,
	numWorkers int,
	repeticiones int,
	referencia ResultadoKMeans,
	tiempoSecuencial float64,
) ResultadoBenchmark {

	tiempos :=
		make([]float64, 0, repeticiones)

	fmt.Println()
	fmt.Println("========================================")

	fmt.Printf(
		" BENCHMARK CONCURRENTE - %d WORKERS\n",
		numWorkers,
	)

	fmt.Println("========================================")

	for ejecucion := 1; ejecucion <= repeticiones; ejecucion++ {

		runtime.GC()

		inicio := time.Now()

		resultado :=
			KMeansConcurrente(
				dataset,
				centroidesIniciales,
				k,
				maxIteraciones,
				tolerancia,
				numWorkers,
			)

		duracion :=
			time.Since(inicio).Seconds()

		tiempos =
			append(
				tiempos,
				duracion,
			)

		correcto :=
			resultadosEquivalentes(
				referencia,
				resultado,
			)

		fmt.Printf(
			"Ejecución %d/%d: %.4f segundos | Resultado correcto: %v\n",
			ejecucion,
			repeticiones,
			duracion,
			correcto,
		)
	}

	media :=
		calcularMedia(tiempos)

	mediaRecortada :=
		calcularMediaRecortada(tiempos)

	// Speedup = T secuencial / T concurrente
	speedup :=
		tiempoSecuencial /
			mediaRecortada

	// Eficiencia = Speedup / número de workers
	eficiencia :=
		speedup /
			float64(numWorkers)

	return ResultadoBenchmark{
		Nombre:         "Concurrente",
		Workers:        numWorkers,
		Tiempos:        tiempos,
		Media:          media,
		MediaRecortada: mediaRecortada,
		Speedup:        speedup,
		Eficiencia:     eficiencia,
	}
}