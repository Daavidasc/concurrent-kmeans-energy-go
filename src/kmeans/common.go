package kmeans

import (
	"math/rand"
)

// Dataset representa los datos que serán procesados por K-Means.
// Los valores se almacenan en un slice plano para reducir el uso de memoria.
type Dataset struct {
	Valores  []float32
	Filas    int
	Features int
}

// ResultadoKMeans representa el resultado final del algoritmo.
type ResultadoKMeans struct {
	Centroides  []float64
	Conteos     []int
	Iteraciones int
	Inercia     float64
}

// InicializarCentroides selecciona K perfiles aleatorios del dataset.
//
// La semilla permite obtener siempre los mismos centroides iniciales,
// lo que será importante para comparar la versión secuencial
// con la versión concurrente.
func InicializarCentroides(
	dataset *Dataset,
	k int,
	seed int64,
) []float64 {

	random := rand.New(rand.NewSource(seed))

	centroides := make(
		[]float64,
		k*dataset.Features,
	)

	// Permite evitar seleccionar dos veces la misma fila.
	filasUsadas := make(map[int]bool)

	cluster := 0

	for cluster < k {

		fila := random.Intn(dataset.Filas)

		if filasUsadas[fila] {
			continue
		}

		filasUsadas[fila] = true

		for feature := 0; feature < dataset.Features; feature++ {

			indiceDato := fila*dataset.Features + feature

			indiceCentroide :=
				cluster*dataset.Features + feature

			centroides[indiceCentroide] =
				float64(dataset.Valores[indiceDato])
		}

		cluster++
	}

	return centroides
}

// CopiarCentroides crea una copia independiente de los centroides.
//
// Esto evita que la versión secuencial y la concurrente
// modifiquen el mismo arreglo durante las pruebas.
func CopiarCentroides(centroides []float64) []float64 {

	copia := make([]float64, len(centroides))

	copy(copia, centroides)

	return copia
}