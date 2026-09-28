package kmeans

import "math"

// KMeansSecuencial ejecuta el algoritmo K-Means
// utilizando un único flujo de ejecución.
func KMeansSecuencial(
	dataset *Dataset,
	centroidesIniciales []float64,
	k int,
	maxIteraciones int,
	tolerancia float64,
) ResultadoKMeans {

	// Copiamos los centroides para no modificar
	// los centroides iniciales originales.
	centroides := CopiarCentroides(
		centroidesIniciales,
	)

	iteraciones := 0

	// ========================================
	// ITERACIONES DE K-MEANS
	// ========================================

	for iteracion := 1; iteracion <= maxIteraciones; iteracion++ {

		// Acumula la suma de las características
		// pertenecientes a cada cluster.
		sumas := make(
			[]float64,
			k*dataset.Features,
		)

		// Cantidad de perfiles asignados a cada cluster.
		conteos := make([]int, k)

		// ========================================
		// PASO 1:
		// ASIGNACIÓN DE PERFILES A CLUSTERS
		// ========================================

		for fila := 0; fila < dataset.Filas; fila++ {

			clusterMasCercano := 0

			menorDistancia := distanciaCuadrada(
				dataset,
				fila,
				centroides,
				0,
			)

			// Comparar con los demás centroides.
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

			// Incrementar cantidad de perfiles
			// asignados al cluster.
			conteos[clusterMasCercano]++

			// Acumular las 48 características.
			for feature := 0; feature < dataset.Features; feature++ {

				indiceDato :=
					fila*dataset.Features + feature

				indiceSuma :=
					clusterMasCercano*dataset.Features + feature

				sumas[indiceSuma] +=
					float64(dataset.Valores[indiceDato])
			}
		}

		// ========================================
		// PASO 2:
		// RECALCULAR CENTROIDES
		// ========================================

		nuevosCentroides := make(
			[]float64,
			len(centroides),
		)

		for cluster := 0; cluster < k; cluster++ {

			// Si ningún perfil fue asignado al cluster,
			// conservamos su centroide anterior.
			if conteos[cluster] == 0 {

				for feature := 0; feature < dataset.Features; feature++ {

					indice :=
						cluster*dataset.Features + feature

					nuevosCentroides[indice] =
						centroides[indice]
				}

				continue
			}

			// Calcular promedio de cada característica.
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

		cambioMaximo := calcularCambioMaximo(
			centroides,
			nuevosCentroides,
		)

		centroides = nuevosCentroides
		iteraciones = iteracion

		// Si los centroides prácticamente ya no cambian,
		// el algoritmo termina.
		if cambioMaximo <= tolerancia {
			break
		}
	}

	// ========================================
	// RESULTADOS FINALES
	// ========================================

	inercia, conteos := calcularInerciaYConteos(
		dataset,
		centroides,
		k,
	)

	return ResultadoKMeans{
		Centroides:  centroides,
		Conteos:     conteos,
		Iteraciones: iteraciones,
		Inercia:     inercia,
	}
}

// distanciaCuadrada calcula la distancia euclidiana al cuadrado
// entre un perfil del dataset y un centroide.
//
// No utilizamos raíz cuadrada porque para determinar cuál
// centroide está más cerca no es necesaria.
func distanciaCuadrada(
	dataset *Dataset,
	fila int,
	centroides []float64,
	cluster int,
) float64 {

	var distancia float64

	for feature := 0; feature < dataset.Features; feature++ {

		indiceDato :=
			fila*dataset.Features + feature

		indiceCentroide :=
			cluster*dataset.Features + feature

		diferencia :=
			float64(dataset.Valores[indiceDato]) -
				centroides[indiceCentroide]

		distancia += diferencia * diferencia
	}

	return distancia
}

// calcularCambioMaximo calcula el mayor cambio producido
// entre los centroides anteriores y los nuevos centroides.
func calcularCambioMaximo(
	anteriores []float64,
	nuevos []float64,
) float64 {

	var cambioMaximo float64

	for i := 0; i < len(anteriores); i++ {

		cambio := math.Abs(
			anteriores[i] - nuevos[i],
		)

		if cambio > cambioMaximo {
			cambioMaximo = cambio
		}
	}

	return cambioMaximo
}

// calcularInerciaYConteos calcula:
//
// 1. La inercia final del modelo.
// 2. La cantidad de perfiles pertenecientes a cada cluster.
func calcularInerciaYConteos(
	dataset *Dataset,
	centroides []float64,
	k int,
) (float64, []int) {

	var inercia float64

	conteos := make([]int, k)

	for fila := 0; fila < dataset.Filas; fila++ {

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

		conteos[clusterMasCercano]++

		// La inercia es la suma de las distancias
		// cuadradas de cada perfil a su centroide.
		inercia += menorDistancia
	}

	return inercia, conteos
}