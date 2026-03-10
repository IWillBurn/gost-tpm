package testing

import (
	"testing"

	"github.com/IWillBurn/gost-go-tpm/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpm2/transport"
	"github.com/stretchr/testify/require"
)

func Test_HasAlgs(t *testing.T) {
	tpm, err := transport.OpenTPM(tpmPath)
	require.NoError(t, err)
	defer tpm.Close()

	testCases := []struct {
		name string
		alg  tpm2.TPMAlgID
	}{
		{
			name: "gost3411_256",
			alg:  tpm2.TPMAlgGOST3411256,
		},
		{
			name: "gost3411_512",
			alg:  tpm2.TPMAlgGOST3411512,
		},
		{
			name: "magma",
			alg:  tpm2.TPMAlgMagma,
		},
		{
			name: "grasshopper",
			alg:  tpm2.TPMAlgGrasshopper,
		},
		{
			name: "gost3410_256",
			alg:  tpm2.TPMAlgGOST3410256,
		},
		{
			name: "gost3410_512",
			alg:  tpm2.TPMAlgGOST3410512,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			found, err := getCapabilityAlg(tpm, testCase.alg)
			require.NoError(t, err)
			require.True(t, found)
		})
	}
}
