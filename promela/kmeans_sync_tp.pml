/* Modelo de sincronizacion del K-Means concurrente (Worker Pool + barrera) - version TP.
 * Mismo diseno que promela/kmeans_sync.pml, con variables de instrumentacion
 * para expresar las propiedades como formulas LTL.                              */
#define STOP 255
#ifndef N_WORKERS
#define N_WORKERS 2
#endif
#ifndef N_JOBS
#define N_JOBS 4
#endif
#ifndef N_ITERS
#define N_ITERS 2
#endif

chan jobs    = [N_JOBS + N_WORKERS] of { byte };
chan results = [N_JOBS] of { byte, byte };

byte centroidVersion = 0;
bool updating  = false;  /* el Master esta escribiendo los centroides        */
byte busy      = 0;      /* workers procesando un bloque en este momento     */
byte resultsIn = 0;      /* resultados recibidos por el Master (iteracion)   */
byte ended     = 0;      /* procesos que llegaron a su fin                   */

proctype Worker(byte workerID)
{
    byte jobID;
    byte localVersion;
    do
    :: jobs ? jobID ->
        if
        :: jobID == STOP -> break
        :: else ->
            busy++;
            assert(!updating);
            localVersion = centroidVersion;
            skip;                                  /* calculo abstraido      */
            assert(!updating);
            assert(localVersion == centroidVersion);
            busy--;
            results ! jobID, workerID
        fi
    od;
    ended++
}

proctype Master()
{
    byte iteration = 0;
    byte jobID;
    byte received;
    byte finishedJob;
    byte workerID;
    byte stopCount;
    do
    :: iteration < N_ITERS ->
        jobID = 0;
        do
        :: jobID < N_JOBS -> jobs ! jobID; jobID++
        :: else -> break
        od;
#ifdef MUT_EARLY_UPDATE
        updating = true; centroidVersion++; updating = false;      /* BUG */
#endif
        received = 0;
        do
#ifdef MUT_OFF_BY_ONE
        :: received < N_JOBS + 1 ->                                /* BUG */
#else
        :: received < N_JOBS ->
#endif
            results ? finishedJob, workerID;
            received++;
            resultsIn++
        :: else -> break
        od;
        assert(received == N_JOBS);
#ifndef MUT_EARLY_UPDATE
        updating = true;
        centroidVersion++;
        updating = false;
#endif
        resultsIn = 0;
        iteration++
    :: else ->
        stopCount = 0;
        do
#ifdef MUT_NO_STOP
        :: stopCount < N_WORKERS - 1 -> jobs ! STOP; stopCount++   /* BUG */
#else
        :: stopCount < N_WORKERS     -> jobs ! STOP; stopCount++
#endif
        :: else -> break
        od;
        break
    od;
    ended++
}

init
{
    byte i = 0;
    atomic {
        do
        :: i < N_WORKERS -> run Worker(i); i++
        :: else -> break
        od;
        run Master()
    }
}

#ifdef LTL
ltl exclusion { [] (updating -> busy == 0) }
ltl barrera   { [] (updating -> resultsIn == N_JOBS) }
ltl termina   { <> (ended == N_WORKERS + 1) }
#endif