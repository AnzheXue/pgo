package gcounter

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/UBC-NSS/pgo/distsys"
	"github.com/UBC-NSS/pgo/distsys/hashmap"
	"github.com/UBC-NSS/pgo/distsys/resources"
	"github.com/UBC-NSS/pgo/distsys/tla"
)

func getNodeMapCtx(self tla.Value, nodeAddrMap *hashmap.HashMap[string], constants []distsys.MPCalContextConfigFn) *distsys.MPCalContext {
	var peers []tla.Value
	for _, nid := range nodeAddrMap.Keys() {
		if !nid.Equal(self) {
			peers = append(peers, nid)
		}
	}
	ctx := distsys.NewMPCalContext(self, ANode, append(constants,
		distsys.EnsureArchetypeRefParam("cntr", resources.NewIncMap(func(index tla.Value) distsys.ArchetypeResource {
			if !index.Equal(self) {
				panic("wrong index")
			}
			return resources.NewCRDT(index, peers, func(index tla.Value) string {
				addr, ok := nodeAddrMap.Get(index)
				if !ok {
					panic(fmt.Errorf("key %v not found in nodeAddrMap %v", index, nodeAddrMap))
				}
				return addr
			}, resources.GCounter{})
		})),
		distsys.EnsureArchetypeRefParam("c", resources.NewDummy()),
	)...)
	return ctx
}

func makeNodeBenchCtx(self tla.Value, nodeAddrMap map[tla.Value]string,
	constants []distsys.MPCalContextConfigFn, outCh chan tla.Value) *distsys.MPCalContext {
	var peers []tla.Value
	for nid := range nodeAddrMap {
		if !nid.Equal(self) {
			peers = append(peers, nid)
		}
	}
	ctx := distsys.NewMPCalContext(self, ANodeBench, append(constants,
		distsys.EnsureArchetypeRefParam("cntr", resources.NewIncMap(func(index tla.Value) distsys.ArchetypeResource {
			if !index.Equal(self) {
				panic("wrong index")
			}
			return resources.NewCRDT(index, peers, func(index tla.Value) string {
				return nodeAddrMap[index]
			}, resources.GCounter{})
		})),
		distsys.EnsureArchetypeRefParam("out", resources.NewOutputChan(outCh)),
	)...)
	return ctx
}

func Bench(t *testing.T, numNodes int, numRounds int) {
	numEvents := numNodes * numRounds * 2

	constants := []distsys.MPCalContextConfigFn{
		distsys.DefineConstantValue("NUM_NODES", tla.MakeNumber(int32(numNodes))),
		distsys.DefineConstantValue("BENCH_NUM_ROUNDS", tla.MakeNumber(int32(numRounds))),
	}
	iface := distsys.NewMPCalContextWithoutArchetype(constants...).IFace()

	nodeAddrMap := make(map[tla.Value]string, numNodes+1)
	for i := 1; i <= numNodes; i++ {
		portNum := 9000 + i
		addr := fmt.Sprintf("localhost:%d", portNum)
		nodeAddrMap[tla.MakeNumber(int32(i))] = addr
	}

	var replicaCtxs []*distsys.MPCalContext
	outCh := make(chan tla.Value, numEvents)
	errs := make(chan error, numNodes)
	for i := 1; i <= numNodes; i++ {
		ctx := makeNodeBenchCtx(tla.MakeNumber(int32(i)), nodeAddrMap, constants, outCh)
		replicaCtxs = append(replicaCtxs, ctx)
		go func() {
			errs <- ctx.Run()
		}()
	}

	starts := make(map[int32]time.Time)
	for i := 0; i < numEvents; i++ {
		resp := <-outCh
		node := resp.ApplyFunction(tla.MakeString("node")).AsNumber()
		event := resp.ApplyFunction(tla.MakeString("event"))
		if event.Equal(IncStart(iface)) {
			starts[node] = time.Now()
		} else if event.Equal(IncFinish(iface)) {
			elapsed := time.Since(starts[node])
			log.Println(node, elapsed)
		}
	}

	for i := 0; i < numNodes; i++ {
		err := <-errs
		if err != nil {
			t.Fatal(err)
		}
	}
}
