package testing

import (
	"fmt"
	"io"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IWillBurn/gost-go-tpm/legacy/tpm2"
)

func Test_HasAlgs(t *testing.T) {

	rwc, err := tpm2.OpenTPM(tpmPath)
	require.NoError(t, err)

	testCases := []struct {
		name string
		alg  tpm2.Algorithm
	}{
		{
			name: "gost3411_256",
			alg:  tpm2.AlgGOST3411_256,
		},
		{
			name: "gost3411_512",
			alg:  tpm2.AlgGOST3411_512,
		},
		{
			name: "magma",
			alg:  tpm2.AlgMagma,
		},
		{
			name: "grasshopper",
			alg:  tpm2.AlgGrasshopper,
		},
		{
			name: "gost3410_256",
			alg:  tpm2.AlgGOST3410_256,
		},
		{
			name: "gost3410_512",
			alg:  tpm2.AlgGOST3410_512,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			passed, err := containsAlg(rwc, testCase.alg)
			require.NoError(t, err)
			require.True(t, passed)
		})
	}
}

func containsAlg(rwc io.ReadWriteCloser, reqAlg tpm2.Algorithm) (bool, error) {
	algorithms, _, err := tpm2.GetCapability(rwc, tpm2.CapabilityAlgs, 1, uint32(reqAlg))

	if err != nil {
		return false, err
	}

	if len(algorithms) == 0 {
		return false, nil
	}

	algInterface := algorithms[0]

	alg, ok := algInterface.(tpm2.AlgorithmDescription)

	if !ok {
		return false, fmt.Errorf("cast to AlgorithmDescription failed")
	}

	if alg.ID != tpm2.Algorithm(reqAlg) {
		return false, nil
	}

	return true, nil
}
