package main

import (
	"fmt"
	"runtime"
	"time"

	"concurrent-kmeans-energy-go/src/kmeans"
)

const (
	// Ruta real de nuestro dataset preparado.
	RutaDataset = "./notebooks/data/kmeans_input.csv"

	// Configuración de K-Means.
	NumFeatures = 48
	K           = 4

	MaxIteraciones = 100
	Tolerancia     = 0.000001

	Seed int64 = 42

	// Para comprobar que todo funciona:
	// Repeticiones = 3
	//
	// Para los resultados finales:
	// Repeticiones = 10
	Repeticiones = 10
)

func main() {

	fmt.Println("========================================")
	fmt.Println(" K-MEANS - CONSUMO ELÉCTRICO")
	fmt.Println(" BENCHMARK SECUENCIAL VS CONCURRENTE")
	fmt.Println("========================================")

	// ========================================
	// PASO 1:
	// INFORMACIÓN DEL EQUIPO
	// ========================================

	fmt.Println()
	fmt.Println("💻 INFORMACIÓN DEL SISTEMA")

	fmt.Printf(
		"CPUs lógicos disponibles: %d\n",
		runtime.NumCPU(),
	)

	fmt.Printf(
		"GOMAXPROCS: %d\n",
		runtime.GOMAXPROCS(0),
	)

	// ========================================
	// PASO 2:
	// CARGAR DATASET
	// ========================================

	fmt.Println()
	fmt.Println("📂 Cargando dataset...")

	inicioCarga := time.Now()

	dataset, err :=
		CargarDataset(
			RutaDataset,
			NumFeatures,
		)

	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	tiempoCarga :=
		time.Since(inicioCarga)

	fmt.Println("✅ Dataset cargado correctamente")

	fmt.Printf(
		"Filas: %d\n",
		dataset.Filas,
	)

	fmt.Printf(
		"Features: %d\n",
		dataset.Features,
	)

	fmt.Printf(
		"Valores almacenados: %d\n",
		len(dataset.Valores),
	)

	fmt.Printf(
		"Tiempo de carga: %v\n",
		tiempoCarga,
	)

	// ========================================
	// PASO 3:
	// CONFIGURACIÓN DEL EXPERIMENTO
	// ========================================

	fmt.Println()
	fmt.Println("⚙️ CONFIGURACIÓN")

	fmt.Printf(
		"K: %d\n",
		K,
	)

	fmt.Printf(
		"Seed: %d\n",
		Seed,
	)

	fmt.Printf(
		"Máximo de iteraciones: %d\n",
		MaxIteraciones,
	)

	fmt.Printf(
		"Tolerancia: %.8f\n",
		Tolerancia,
	)

	fmt.Printf(
		"Repeticiones por configuración: %d\n",
		Repeticiones,
	)

	// ========================================
	// PASO 4:
	// INICIALIZAR CENTROIDES
	// ========================================

	fmt.Println()
	fmt.Println("🎯 Inicializando centroides...")

	centroidesIniciales :=
		kmeans.InicializarCentroides(
			dataset,
			K,
			Seed,
		)

	fmt.Println("✅ Centroides inicializados")

	// ========================================
	// PASO 5:
	// BENCHMARK SECUENCIAL
	// ========================================

	benchmarkSecuencial,
		resultadoReferencia :=
		kmeans.BenchmarkSecuencial(
			dataset,
			centroidesIniciales,
			K,
			MaxIteraciones,
			Tolerancia,
			Repeticiones,
		)

	// ========================================
	// VALIDACIÓN DEL MODELO SECUENCIAL
	// ========================================

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(" RESULTADO DE REFERENCIA")
	fmt.Println("========================================")

	fmt.Printf(
		"Iteraciones: %d\n",
		resultadoReferencia.Iteraciones,
	)

	fmt.Printf(
		"Inercia: %.8f\n",
		resultadoReferencia.Inercia,
	)

	fmt.Println()
	fmt.Println("Distribución por cluster:")

	for cluster, cantidad :=
		range resultadoReferencia.Conteos {

		porcentaje :=
			float64(cantidad) /
				float64(dataset.Filas) *
				100

		fmt.Printf(
			"Cluster %d: %d perfiles (%.2f%%)\n",
			cluster,
			cantidad,
			porcentaje,
		)
	}

	// ========================================
	// PASO 6:
	// CONFIGURACIONES CONCURRENTES
	// ========================================

	configuracionesWorkers :=
		[]int{
			2,
			4,
			8,
			16,
		}

	resultados :=
		[]kmeans.ResultadoBenchmark{
			benchmarkSecuencial,
		}

	// ========================================
	// PASO 7:
	// BENCHMARK CONCURRENTE
	// ========================================

	for _, workers :=
		range configuracionesWorkers {

		benchmarkConcurrente :=
			kmeans.BenchmarkConcurrente(
				dataset,
				centroidesIniciales,
				K,
				MaxIteraciones,
				Tolerancia,
				workers,
				Repeticiones,
				resultadoReferencia,
				benchmarkSecuencial.MediaRecortada,
			)

		resultados =
			append(
				resultados,
				benchmarkConcurrente,
			)
	}

	// ========================================
	// PASO 8:
	// RESULTADOS FINALES
	// ========================================

	fmt.Println()
	fmt.Println("======================================================================")
	fmt.Println(" RESULTADOS DEL BENCHMARK")
	fmt.Println("======================================================================")

	fmt.Printf(
		"%-15s %-10s %-12s %-14s %-10s %-12s\n",
		"Modelo",
		"Workers",
		"Media",
		"Media Rec.",
		"Speedup",
		"Eficiencia",
	)

	fmt.Println("----------------------------------------------------------------------")

	for _, resultado :=
		range resultados {

		fmt.Printf(
			"%-15s %-10d %-12.4f %-14.4f %-10.2fx %-11.2f%%\n",
			resultado.Nombre,
			resultado.Workers,
			resultado.Media,
			resultado.MediaRecortada,
			resultado.Speedup,
			resultado.Eficiencia*100,
		)
	}

	// ========================================
	// PASO 9:
	// MOSTRAR TODAS LAS EJECUCIONES
	// ========================================

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(" TIEMPOS INDIVIDUALES")
	fmt.Println("========================================")

	for _, resultado :=
		range resultados {

		fmt.Printf(
			"\n%s - %d worker(s):\n",
			resultado.Nombre,
			resultado.Workers,
		)

		for i, tiempo :=
			range resultado.Tiempos {

			fmt.Printf(
				"  Ejecución %d: %.4f segundos\n",
				i+1,
				tiempo,
			)
		}
	}

	fmt.Println()
	fmt.Println("🏁 Benchmark finalizado")
}