package shopcart

import (
	"log"

	"github.com/DistCompiler/pgo/distsys"
	"github.com/DistCompiler/pgo/distsys/resources"
	"github.com/DistCompiler/pgo/distsys/tla"
	"github.com/DistCompiler/pgo/systems/shopcart/configs"
)

func makeConstants(c configs.Root) []distsys.MPCalContextConfigFn {
	constants := []distsys.MPCalContextConfigFn{
		distsys.DefineConstantValue("NumNodes", tla.MakeNumber(int32(len(c.Peers)))),
		distsys.DefineConstantValue("ElemSet", tla.MakeSet()),
		distsys.DefineConstantValue("BenchNumRounds", tla.MakeNumber(int32(c.NumRounds))),
	}
	return constants
}

func newNodeBenchCtx(self tla.Value, c configs.Root, outCh chan tla.Value) *distsys.MPCalContext {
	constants := makeConstants(c)

	toMap := func(res distsys.ArchetypeResource) distsys.ArchetypeResource {
		return resources.NewIncMap(func(index tla.Value) distsys.ArchetypeResource {
			if index.Equal(self) {
				return res
			}
			panic("wrong index")
		})
	}

	var peers []tla.Value
	for peerId := range c.Peers {
		peers = append(peers, tla.MakeNumber(int32(peerId)))
	}
	addrMapper := func(idx tla.Value) string {
		idxNum := int(idx.AsNumber())
		addr, ok := c.Peers[idxNum]
		if !ok {
			panic("peer not found")
		}
		return addr
	}

	crdt := resources.NewCRDT(self, peers, addrMapper, resources.LWWSet{},
		resources.WithCRDTBroadcastInterval(c.BroadcastInterval),
		resources.WithCRDTSendTimeout(c.SendTimeout),
		resources.WithCRDTDialTimeout(c.DialTimeout),
	)
	out := resources.NewOutputChan(outCh)
	cDummy := resources.NewDummy(resources.WithDummyValue(tla.MakeSet()))

	ctx := distsys.NewMPCalContext(self, ANodeBench,
		distsys.EnsureMPCalContextConfigs(constants...),
		distsys.EnsureArchetypeRefParam("crdt", toMap(crdt)),
		distsys.EnsureArchetypeRefParam("out", out),
		distsys.EnsureArchetypeRefParam("c", cDummy),
	)
	return ctx
}

type Event int

const (
	AddStartEvent Event = iota + 1
	AddFinishEvent
)

func (e Event) String() string {
	switch e {
	case AddStartEvent:
		return "AddStart"
	case AddFinishEvent:
		return "AddFinish"
	default:
		return "unknown"
	}
}

type Node struct {
	Id     int
	Config configs.Root

	ctx   *distsys.MPCalContext
	ch    chan Event
	tlaCh chan tla.Value
}

func NewNode(id int, c configs.Root, ch chan Event) *Node {
	self := tla.MakeNumber(int32(id))

	tlaCh := make(chan tla.Value, 100)
	ctx := newNodeBenchCtx(self, c, tlaCh)
	return &Node{
		Id:     id,
		Config: c,
		ctx:    ctx,
		ch:     ch,
		tlaCh:  tlaCh,
	}
}

func (n *Node) Run() error {
	iface := distsys.NewMPCalContextWithoutArchetype().IFace()
	numEvents := (n.Config.NumRounds - 1) * 2

	errCh := make(chan error)
	go func() {
		err := n.ctx.Run()
		log.Printf("node %v done, err = %v", n.Id, err)
		if err != nil {
			log.Fatal(err)
		}
		errCh <- err
	}()

	for i := 0; i < numEvents; i++ {
		resp := <-n.tlaCh
		// log.Println(resp)
		event := resp.ApplyFunction(tla.MakeString("event"))
		if event.Equal(AddStart(iface)) {
			n.ch <- AddStartEvent
		} else if event.Equal(AddFinish(iface)) {
			n.ch <- AddFinishEvent
		}
	}

	err := <-errCh
	return err
}

func (n *Node) Close() error {
	n.ctx.Stop()
	return nil
}
