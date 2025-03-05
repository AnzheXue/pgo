package test

import (
	"testing"

	"github.com/DistCompiler/pgo/distsys"
	"github.com/DistCompiler/pgo/distsys/resources"
	"github.com/DistCompiler/pgo/distsys/tla"
)

func TestCounter(t *testing.T) {
	outChan := make(chan tla.Value, 1)

	ctx := distsys.NewMPCalContext(tla.MakeString("self"), Counter,
		distsys.EnsureArchetypeRefParam("out", resources.NewOutputChan(outChan)))

	err := ctx.Run()
	if err != nil {
		panic(err)
	}

	outVal := <-outChan
	close(outChan) // everything is sync in this test, but close the channel anyway to catch anything weird
	if !outVal.Equal(tla.MakeNumber(1)) {
		t.Errorf("incrementation result %v was not equal to expected value 1", outVal)
	}
}
