package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"concurrent-kmeans-energy-go/src/kmeans"
)

// CargarDataset lee el archivo CSV preparado
// para K-Means.
func CargarDataset(
	ruta string,
	numFeatures int,
) (*kmeans.Dataset, error) {

	archivo, err :=
		os.Open(ruta)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudo abrir el dataset: %w",
				err,
			)
	}

	defer archivo.Close()

	// Buffer de 4 MB para mejorar la lectura.
	buffer :=
		bufio.NewReaderSize(
			archivo,
			4*1024*1024,
		)

	reader :=
		csv.NewReader(buffer)

	reader.ReuseRecord = true

	// ========================================
	// LEER CABECERA
	// ========================================

	header, err :=
		reader.Read()

	if err != nil {
		return nil,
			fmt.Errorf(
				"error leyendo cabecera: %w",
				err,
			)
	}

	if len(header) != numFeatures {

		return nil,
			fmt.Errorf(
				"se esperaban %d columnas, pero se encontraron %d",
				numFeatures,
				len(header),
			)
	}

	// Cantidad aproximada conocida después
	// de la preparación del dataset.
	const filasEsperadas = 3454475

	valores :=
		make(
			[]float32,
			0,
			filasEsperadas*numFeatures,
		)

	filas := 0

	// ========================================
	// LEER REGISTROS
	// ========================================

	for {

		registro, err :=
			reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {

			return nil,
				fmt.Errorf(
					"error leyendo fila %d: %w",
					filas+1,
					err,
				)
		}

		if len(registro) != numFeatures {

			return nil,
				fmt.Errorf(
					"fila %d tiene %d columnas, se esperaban %d",
					filas+1,
					len(registro),
					numFeatures,
				)
		}

		for columna, texto :=
			range registro {

			valor, err :=
				strconv.ParseFloat(
					texto,
					32,
				)

			if err != nil {

				return nil,
					fmt.Errorf(
						"valor inválido en fila %d columna %d: %w",
						filas+1,
						columna+1,
						err,
					)
			}

			valores =
				append(
					valores,
					float32(valor),
				)
		}

		filas++
	}

	dataset :=
		&kmeans.Dataset{
			Valores:  valores,
			Filas:    filas,
			Features: numFeatures,
		}

	return dataset, nil
}