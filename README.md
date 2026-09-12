# Segmentación de Hogares por Patrones de Consumo Eléctrico con K-Means Concurrente en Go

## Integrantes
- Fuentes Rivera Onofre, Marco Antonio 
- Davila Mundo, Alexis
- Serrudo Caina, David Andre

## Objetivo de Desarrollo Sostenible
**ODS 7: Energía asequible y no contaminante**

## Descripción del proyecto
Este proyecto analiza la aplicación de programación concurrente (usando goroutines en Go) al algoritmo de Machine Learning **K-Means**, con el fin de segmentar hogares según sus patrones de consumo eléctrico. La segmentación permite identificar comportamientos de consumo (picos sincronizados, consumo estable, consumo ineficiente) que sirven para apoyar decisiones de planificación energética: balanceo de carga en la red, tarifas diferenciadas y campañas de eficiencia energética dirigidas.

## Dataset
**Smart Meters in London** — lecturas de consumo eléctrico cada 30 minutos de 5,567 hogares de Londres (nov. 2011 – feb. 2014), recolectadas en el proyecto Low Carbon London de UK Power Networks.
- Fuente: [Kaggle - jeanmidev/smart-meters-in-london](https://www.kaggle.com/datasets/jeanmidev/smart-meters-in-london)
- Registros: +1,000,000 (half-hourly dataset)
- Variables principales: `LCLid` (hogar), `DateTime`, `KWH/hh` (consumo), `stdorToU` (tipo de tarifa), `Acorn`/`Acorn_grouped` (clasificación socioeconómica)

## Modelo de Machine Learning
**K-Means** (clustering no supervisado) — algoritmo "embarrassingly parallel", ideal para demostrar patrones de concurrencia tipo Worker Pool (fan-out / fan-in) en Go.

## Estructura del repositorio

├── docs/ # Informes (PC1, PC2, TP)
├── notebooks/ # EDA y limpieza del dataset
├── data/ # Solo referencia al dataset (no se sube por peso)
├── src/ # Código Go (secuencial y concurrente) — desde PC2
└── README.md


## Estado del proyecto
- [x] PC1: Investigación bibliográfica, definición de caso de uso, selección y limpieza del dataset
- [ ] PC2: Modelado en Promela, implementación Go (secuencial vs. concurrente), cálculo de speedup
- [ ] TP: Verificación formal en Spin, informe de análisis con IA, conclusiones finales

## Papers de referencia
1. Paul, D. C., Quadir, H., Bellatreche, L., & Ordonez, C. (2026). *A generic algorithm integrating parallel processing and block-based I/O for large-scale AI analytics*. Data & Knowledge Engineering, 167.
2. Djafri, L. (2022). *Dynamic Distributed and Parallel Machine Learning algorithms for big data mining processing*. Data Technologies and Applications, 56(4), 558–601.
3. Jeon, S., Lee, S., & Lee, I. (2023). *Hybrid in-network computing and distributed learning for large-scale data processing*. Computer Networks, 226.
