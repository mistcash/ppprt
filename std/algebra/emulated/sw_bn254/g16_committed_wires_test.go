package sw_bn254

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/frontend/cs/scs"
)

// TestGroth16SimulationCommittedWires logs the wires the Groth16Simulation
// circuit's commitments cover, per constraint system. It is the measure
// mistcash/polyring's paper/ benchmark reports, so the two can be compared;
// the "nb commitments" of BenchmarkGroth16Simulation instead sums the values
// passed to api.Commit, duplicates included. Test-only: the measured code is
// unchanged.
func TestGroth16SimulationCommittedWires(t *testing.T) {
	field := ecc.BN254.ScalarField()
	newCircuit := func() frontend.Circuit {
		return &Groth16Simulation{In1G2: NewG2AffineFixedPlaceholder(), In2G2: NewG2AffineFixedPlaceholder()}
	}

	ccs, err := frontend.Compile(field, scs.NewBuilder, newCircuit())
	if err != nil {
		t.Fatal(err)
	}
	nbSCS := 0
	for _, cm := range ccs.GetCommitments().(constraint.PlonkCommitments) {
		nbSCS += len(cm.Committed)
	}

	ccs, err = frontend.Compile(field, r1cs.NewBuilder, newCircuit())
	if err != nil {
		t.Fatal(err)
	}
	nbR1CS := 0
	for _, cm := range ccs.GetCommitments().(constraint.Groth16Commitments) {
		nbR1CS += len(cm.PrivateCommitted) + len(cm.PublicAndCommitmentCommitted)
	}

	t.Logf("committed wires: scs %d, r1cs %d", nbSCS, nbR1CS)
}
