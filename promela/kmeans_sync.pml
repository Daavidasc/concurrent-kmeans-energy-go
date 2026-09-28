#define N_WORKERS 2
#define N_JOBS 4
#define N_ITERS 2
#define STOP 255

/* 
* Canal compartido de trabajos.
* Representa los bloques de perfiles que el Master
* distribuye entre los workers.
*/ 
chan jobs = [N_JOBS + N_WORKERS] of { byte };

/* 
* Canal mediante el cual los workers envían
* sus resultados parciales al Master.
* 
* Se envía:
* - ID del trabajo
* - ID del worker
*/ 
chan results = [N_JOBS] of { byte,byte };

/* 
* Versión actual de los centroides.
*/ 
byte centroidVersion = 0;

/* 
* Indica si el Master está actualizando centroides.
*/ 
bool updating = false;


/* 
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
* WORKER
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
*/ 
proctype Worker(byte workerID)
{
	byte jobID;
	byte localVersion;
	
	do
		
		/* Esperar un trabajo*/ 
	:: jobs?jobID -> 
		
		if
			
			/* Señal para terminar*/ 
		:: jobID == STOP -> 
			break
			
			/* Procesamiento normal*/ 
		:: else -> 
			
			/* 
			* El Master no debe estar actualizando
			* centroides cuando comienza el trabajo.
			*/ 
			assert(!updating);
			
			/* 
			* Guardamos la versión actual
			* de los centroides.
			*/ 
			localVersion = centroidVersion;
			
			/* 
			* Aquí se abstrae el cálculo real:
			* 
			* - distancia euclidiana
			* - asignación al cluster
			* - sumas locales
			* - conteos locales
			*/ 
			skip;
			
			/* 
			* Comprobar que el Master no modificó
			* centroides mientras el worker procesaba.
			*/ 
			assert(!updating);
			
			assert(
			localVersion == centroidVersion
			);
			
			/* 
			* Informar al Master que el bloque
			* terminó de procesarse.
			*/ 
			results!jobID,workerID
			
		fi
		
	od
}


/* 
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
* MASTER
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
*/ 
proctype Master()
{
	byte iteration = 0;
	
	byte jobID;
	byte received;
	
	byte finishedJob;
	byte workerID;
	
	byte stopCount;
	
	do
		
		/* 
		* Ejecutar iteraciones de K - Means.
		*/ 
	:: iteration < N_ITERS -> 
		
		/* 
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* PASO 1:
		* DISTRIBUIR TRABAJOS
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		*/ 
		jobID = 0;
		
		do
			
		:: jobID < N_JOBS -> 
			
			jobs!jobID;
			
			jobID++
			
		:: else -> 
			break
			
		od;
		
		
		/* 
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* PASO 2:
		* BARRERA DE SINCRONIZACIÓN
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* 
		* El Master debe esperar todos los
		* resultados antes de actualizar centroides.
		*/ 
		received = 0;
		
		do
			
		:: received < N_JOBS -> 
			
			results?finishedJob,workerID;
			
			received++
			
		:: else -> 
			break
			
		od;
		
		
		/* 
		* Todos los trabajos deben haber terminado.
		*/ 
		assert(
		received == N_JOBS
		);
		
		
		/* 
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* PASO 3:
		* ACTUALIZAR CENTROIDES
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* 
		* Solo el Master realiza esta operación.
		*/ 
		atomic {
			
			updating = true;
			
			centroidVersion++;
			
			updating = false
		};
		
		
		/* 
		* Siguiente iteración.
		*/ 
		iteration++
		
		
		/* 
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		* FIN DEL ALGORITMO
		* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
		*/ 
	:: else -> 
		
		/* 
		* Enviar una señal STOP por cada Worker.
		*/ 
		stopCount = 0;
		
		do
			
		:: stopCount < N_WORKERS -> 
			
			jobs!STOP;
			
			stopCount++
			
		:: else -> 
			break
			
		od;
		
		break
		
	od
}


/* 
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
* INICIALIZACIÓN
* ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  ==  == 
*/ 
init
{
	run Worker(0);
	
	run Worker(1);
	
	run Master();
}