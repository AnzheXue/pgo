---- MODULE ConcurrentQueueAPI_TTrace_1762643810 ----
EXTENDS Sequences, TLCExt, ConcurrentQueueAPI_TEConstants, Toolbox, Naturals, TLC, ConcurrentQueueAPI

_expression ==
    LET ConcurrentQueueAPI_TEExpression == INSTANCE ConcurrentQueueAPI_TEExpression
    IN ConcurrentQueueAPI_TEExpression!expression
----

_trace ==
    LET ConcurrentQueueAPI_TETrace == INSTANCE ConcurrentQueueAPI_TETrace
    IN ConcurrentQueueAPI_TETrace!trace
----

_inv ==
    ~(
        TLCGet("level") = Len(_TETrace)
        /\
        enqueued = (2)
        /\
        queues = ((t1 :> <<e1>> @@ t2 :> <<e1>> @@ t3 :> <<>> @@ t4 :> <<>>))
        /\
        dequeued = (0)
    )
----

_init ==
    /\ queues = _TETrace[1].queues
    /\ enqueued = _TETrace[1].enqueued
    /\ dequeued = _TETrace[1].dequeued
----

_next ==
    /\ \E i,j \in DOMAIN _TETrace:
        /\ \/ /\ j = i + 1
              /\ i = TLCGet("level")
        /\ queues  = _TETrace[i].queues
        /\ queues' = _TETrace[j].queues
        /\ enqueued  = _TETrace[i].enqueued
        /\ enqueued' = _TETrace[j].enqueued
        /\ dequeued  = _TETrace[i].dequeued
        /\ dequeued' = _TETrace[j].dequeued

\* Uncomment the ASSUME below to write the states of the error trace
\* to the given file in Json format. Note that you can pass any tuple
\* to `JsonSerialize`. For example, a sub-sequence of _TETrace.
    \* ASSUME
    \*     LET J == INSTANCE Json
    \*         IN J!JsonSerialize("ConcurrentQueueAPI_TTrace_1762643810.json", _TETrace)

=============================================================================

 Note that you can extract this module `ConcurrentQueueAPI_TEExpression`
  to a dedicated file to reuse `expression` (the module in the 
  dedicated `ConcurrentQueueAPI_TEExpression.tla` file takes precedence 
  over the module `ConcurrentQueueAPI_TEExpression` below).

---- MODULE ConcurrentQueueAPI_TEExpression ----
EXTENDS Sequences, TLCExt, ConcurrentQueueAPI_TEConstants, Toolbox, Naturals, TLC, ConcurrentQueueAPI

expression == 
    [
        \* To hide variables of the `ConcurrentQueueAPI` spec from the error trace,
        \* remove the variables below.  The trace will be written in the order
        \* of the fields of this record.
        queues |-> queues
        ,enqueued |-> enqueued
        ,dequeued |-> dequeued
        
        \* Put additional constant-, state-, and action-level expressions here:
        \* ,_stateNumber |-> _TEPosition
        \* ,_queuesUnchanged |-> queues = queues'
        
        \* Format the `queues` variable as Json value.
        \* ,_queuesJson |->
        \*     LET J == INSTANCE Json
        \*     IN J!ToJson(queues)
        
        \* Lastly, you may build expressions over arbitrary sets of states by
        \* leveraging the _TETrace operator.  For example, this is how to
        \* count the number of times a spec variable changed up to the current
        \* state in the trace.
        \* ,_queuesModCount |->
        \*     LET F[s \in DOMAIN _TETrace] ==
        \*         IF s = 1 THEN 0
        \*         ELSE IF _TETrace[s].queues # _TETrace[s-1].queues
        \*             THEN 1 + F[s-1] ELSE F[s-1]
        \*     IN F[_TEPosition - 1]
    ]

=============================================================================



Parsing and semantic processing can take forever if the trace below is long.
 In this case, it is advised to uncomment the module below to deserialize the
 trace from a generated binary file.

\*
\*---- MODULE ConcurrentQueueAPI_TETrace ----
\*EXTENDS IOUtils, ConcurrentQueueAPI_TEConstants, TLC, ConcurrentQueueAPI
\*
\*trace == IODeserialize("ConcurrentQueueAPI_TTrace_1762643810.bin", TRUE)
\*
\*=============================================================================
\*

---- MODULE ConcurrentQueueAPI_TETrace ----
EXTENDS ConcurrentQueueAPI_TEConstants, TLC, ConcurrentQueueAPI

trace == 
    <<
    ([enqueued |-> 0,queues |-> (t1 :> <<>> @@ t2 :> <<>> @@ t3 :> <<>> @@ t4 :> <<>>),dequeued |-> 0]),
    ([enqueued |-> 1,queues |-> (t1 :> <<e1>> @@ t2 :> <<>> @@ t3 :> <<>> @@ t4 :> <<>>),dequeued |-> 0]),
    ([enqueued |-> 2,queues |-> (t1 :> <<e1>> @@ t2 :> <<e1>> @@ t3 :> <<>> @@ t4 :> <<>>),dequeued |-> 0])
    >>
----


=============================================================================

---- MODULE ConcurrentQueueAPI_TEConstants ----
EXTENDS ConcurrentQueueAPI

CONSTANTS e1, e2, e3, e4, t1, t2, t3, t4

=============================================================================

---- CONFIG ConcurrentQueueAPI_TTrace_1762643810 ----
CONSTANTS
    Elements = { e1 , e2 , e3 , e4 }
    MaxElements = 5
    Threads = { t1 , t2 , t3 , t4 }
    MaxBulkSize = 3
    e3 = e3
    t2 = t2
    e2 = e2
    t3 = t3
    e4 = e4
    e1 = e1
    t1 = t1
    t4 = t4

INVARIANT
    _inv

CHECK_DEADLOCK
    \* CHECK_DEADLOCK off because of PROPERTY or INVARIANT above.
    FALSE

INIT
    _init

NEXT
    _next

CONSTANT
    _TETrace <- _trace

ALIAS
    _expression
=============================================================================
\* Generated on Sat Nov 08 18:16:52 EST 2025