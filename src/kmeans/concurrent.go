package kmeans

import "sync"

// TrabajoKMeans representa un bloque de filas
// que será procesado por un worker.
type TrabajoKMeans struct {
	ID     int
	Inicio int
	Fin    int
}

// ResultadoParcial contiene los resultados obtenidos
// por un worker para un bloque del dataset.
type ResultadoParcial struct {
	ID       int
	WorkerID int
	Sumas    []float64
	Conteos  []int
}

// ResultadoMetricas contiene la información necesaria
// para calcular la inercia final.
type ResultadoMetricas struct {
	ID       int
	Inercia  float64
	Conteos  []int
}

// workerAsignacion procesa los bloques recibidos
// desde el canal de trabajos.
func workerAsignacion(
	workerID int,
	dataset *Dataset,
	centroides []float64,
	k int,
	trabajos <-chan TrabajoKMeans,
	resultados chan<- ResultadoParcial,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for trabajo := range trabajos {

		sumasLocales := make(
			[]float64,
			k*dataset.Features,
		)

		conteosLocales := make([]int, k)

		for fila := trabajo.Inicio; fila < trabajo.Fin; fila++ {

			clusterMasCercano := 0

			menorDistancia := distanciaCuadrada(
				dataset,
				fila,
				centroides,
				0,
			)

			for cluster := 1; cluster < k; cluster++ {

				distancia := distanciaCuadrada(
					dataset,
					fila,
					centroides,
					cluster,
				)

				if distancia < menorDistancia {
					menorDistancia = distancia
					clusterMasCercano = cluster
				}
			}

			conteosLocales[clusterMasCercano]++

			for feature := 0; feature < dataset.Features; feature++ {

				indiceDato :=
					fila*dataset.Features + feature

				indiceSuma :=
					clusterMasCercano*dataset.Features + feature

				sumasLocales[indiceSuma] +=
					float64(dataset.Valores[indiceDato])
			}
		}

		resultados <- ResultadoParcial{
			ID:       trabajo.ID,
			WorkerID: workerID,
			Sumas:    sumasLocales,
			Conteos:  conteosLocales,
		}
	}
}

// procesarAsignacionConcurrente ejecuta la etapa de asignación
// utilizando un Worker Pool.
func procesarAsignacionConcurrente(
	dataset *Dataset,
	centroides []float64,
	k int,
	numWorkers int,
) ([]float64, []int) {

	if numWorkers < 1 {
		numWorkers = 1
	}

	// Creamos varios bloques por worker para distribuir
	// mejor la carga de trabajo.
	numBloques := numWorkers * 4

	tamanoBloque :=
		(dataset.Filas + numBloques - 1) / numBloques

	totalTrabajos :=
		(dataset.Filas + tamanoBloque - 1) / tamanoBloque

	trabajosChannel :=
		make(chan TrabajoKMeans, numWorkers)

	resultadosChannel :=
		make(chan ResultadoParcial, totalTrabajos)

	var wg sync.WaitGroup

	// ========================================
	// PASO 1: Iniciar Worker Pool
	// ========================================

	for workerID := 1; workerID <= numWorkers; workerID++ {

		wg.Add(1)

		go workerAsignacion(
			workerID,
			dataset,
			centroides,
			k,
			trabajosChannel,
			resultadosChannel,
			&wg,
		)
	}

	// ========================================
	// PASO 2: Enviar bloques de trabajo
	// ========================================

	go func() {

		id := 0

		for inicio := 0; inicio < dataset.Filas; inicio += tamanoBloque {

			fin := inicio + tamanoBloque

			if fin > dataset.Filas {
				fin = dataset.Filas
			}

			trabajosChannel <- TrabajoKMeans{
				ID:     id,
				Inicio: inicio,
				Fin:    fin,
			}

			id++
		}

		close(trabajosChannel)
	}()

	// ========================================
	// PASO 3: Esperar que terminen los workers
	// ========================================

	go func() {
		wg.Wait()
		close(resultadosChannel)
	}()

	// Guardamos los resultados según el ID del bloque.
	// De esta forma la agregación se realiza siempre
	// en el mismo orden.
	parciales := make(
		[]ResultadoParcial,
		totalTrabajos,
	)

	for resultado := range resultadosChannel {
		parciales[resultado.ID] = resultado
	}

	// ========================================
	// PASO 4: Combinar resultados
	// ========================================

	sumasTotales := make(
		[]float64,
		k*dataset.Features,
	)

	conteosTotales := make([]int, k)

	for id := 0; id < totalTrabajos; id++ {

		resultado := parciales[id]

		for i := 0; i < len(sumasTotales); i++ {
			sumasTotales[i] += resultado.Sumas[i]
		}

		for cluster := 0; cluster < k; cluster++ {
			conteosTotales[cluster] +=
				resultado.Conteos[cluster]
		}
	}

	return sumasTotales, conteosTotales
}

// KMeansConcurrente ejecuta K-Means utilizando
// un Worker Pool con múltiples goroutines.
func KMeansConcurrente(
	dataset *Dataset,
	centroidesIniciales []float64,
	k int,
	maxIteraciones int,
	tolerancia float64,
	numWorkers int,
) ResultadoKMeans {

	centroides :=
		CopiarCentroides(centroidesIniciales)

	iteraciones := 0

	for iteracion := 1; iteracion <= maxIteraciones; iteracion++ {

		// ========================================
		// PASO 1:
		// ASIGNACIÓN CONCURRENTE
		// ========================================

		sumas, conteos :=
			procesarAsignacionConcurrente(
				dataset,
				centroides,
				k,
				numWorkers,
			)

		// ========================================
		// PASO 2:
		// RECALCULAR CENTROIDES
		// ========================================

		nuevosCentroides :=
			make([]float64, len(centroides))

		for cluster := 0; cluster < k; cluster++ {

			if conteos[cluster] == 0 {

				for feature := 0; feature < dataset.Features; feature++ {

					indice :=
						cluster*dataset.Features + feature

					nuevosCentroides[indice] =
						centroides[indice]
				}

				continue
			}

			for feature := 0; feature < dataset.Features; feature++ {

				indice :=
					cluster*dataset.Features + feature

				nuevosCentroides[indice] =
					sumas[indice] /
						float64(conteos[cluster])
			}
		}

		// ========================================
		// PASO 3:
		// VERIFICAR CONVERGENCIA
		// ========================================

		cambioMaximo :=
			calcularCambioMaximo(
				centroides,
				nuevosCentroides,
			)

		centroides = nuevosCentroides
		iteraciones = iteracion

		if cambioMaximo <= tolerancia {
			break
		}
	}

	// Calcular inercia y distribución final.
	inercia, conteos :=
		calcularMetricasConcurrentes(
			dataset,
			centroides,
			k,
			numWorkers,
		)

	return ResultadoKMeans{
		Centroides:  centroides,
		Conteos:     conteos,
		Iteraciones: iteraciones,
		Inercia:     inercia,
	}
}

// workerMetricas calcula la inercia y los conteos
// finales para un bloque de perfiles.
func workerMetricas(
	dataset *Dataset,
	centroides []float64,
	k int,
	trabajos <-chan TrabajoKMeans,
	resultados chan<- ResultadoMetricas,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for trabajo := range trabajos {

		var inerciaLocal float64

		conteosLocales := make([]int, k)

		for fila := trabajo.Inicio; fila < trabajo.Fin; fila++ {

			clusterMasCercano := 0

			menorDistancia := distanciaCuadrada(
				dataset,
				fila,
				centroides,
				0,
			)

			for cluster := 1; cluster < k; cluster++ {

				distancia := distanciaCuadrada(
					dataset,
					fila,
					centroides,
					cluster,
				)

				if distancia < menorDistancia {
					menorDistancia = distancia
					clusterMasCercano = cluster
				}
			}

			conteosLocales[clusterMasCercano]++

			inerciaLocal += menorDistancia
		}

		resultados <- ResultadoMetricas{
			ID:      trabajo.ID,
			Inercia: inerciaLocal,
			Conteos: conteosLocales,
		}
	}
}

// calcularMetricasConcurrentes obtiene la inercia
// y la distribución final utilizando workers.
func calcularMetricasConcurrentes(
	dataset *Dataset,
	centroides []float64,
	k int,
	numWorkers int,
) (float64, []int) {

	numBloques := numWorkers * 4

	tamanoBloque :=
		(dataset.Filas + numBloques - 1) / numBloques

	totalTrabajos :=
		(dataset.Filas + tamanoBloque - 1) / tamanoBloque

	trabajosChannel :=
		make(chan TrabajoKMeans, numWorkers)

	resultadosChannel :=
		make(chan ResultadoMetricas, totalTrabajos)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {

		wg.Add(1)

		go workerMetricas(
			dataset,
			centroides,
			k,
			trabajosChannel,
			resultadosChannel,
			&wg,
		)
	}

	go func() {

		id := 0

		for inicio := 0; inicio < dataset.Filas; inicio += tamanoBloque {

			fin := inicio + tamanoBloque

			if fin > dataset.Filas {
				fin = dataset.Filas
			}

			trabajosChannel <- TrabajoKMeans{
				ID:     id,
				Inicio: inicio,
				Fin:    fin,
			}

			id++
		}

		close(trabajosChannel)
	}()

	go func() {
		wg.Wait()
		close(resultadosChannel)
	}()

	resultadosOrdenados :=
		make([]ResultadoMetricas, totalTrabajos)

	for resultado := range resultadosChannel {
		resultadosOrdenados[resultado.ID] = resultado
	}

	var inerciaTotal float64

	conteosTotales := make([]int, k)

	for id := 0; id < totalTrabajos; id++ {

		resultado := resultadosOrdenados[id]

		inerciaTotal += resultado.Inercia

		for cluster := 0; cluster < k; cluster++ {
			conteosTotales[cluster] +=
				resultado.Conteos[cluster]
		}
	}

	return inerciaTotal, conteosTotales
}