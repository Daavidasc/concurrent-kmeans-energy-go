package kmeans

import (
	"testing"
	"time"
)

/*
Este archivo contiene pruebas unitarias orientadas a validar casos límite
y escenarios excepcionales del algoritmo K-Means.

Las pruebas verifican que el sistema gestione correctamente situaciones
como la inicialización de centroides cuando el valor de K es mayor que la
cantidad de filas disponibles en el dataset, así como el procesamiento de
datasets vacíos. El objetivo es garantizar la estabilidad del algoritmo,
evitando bloqueos, ciclos infinitos o errores inesperados (panic) durante
la ejecución, especialmente en entornos concurrentes.
*/

func TestKMayorQueFilas(t *testing.T) {
	ds := &Dataset{Valores: make([]float32, 3*48), Filas: 3, Features: 48}
	done := make(chan struct{})
	go func() { InicializarCentroides(ds, 4, 42); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("InicializarCentroides no termina con K > filas")
	}
}

func TestCeroFilas(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic con 0 filas: %v", r)
		}
	}()
	ds := &Dataset{Features: 48}
	procesarAsignacionConcurrente(ds, make([]float64, 4*48), 4, 4)
}
