package testing

import (
	"fmt"
	"io"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IWillBurn/gost-go-tpm/legacy/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpmutil"
)

func Test_FakeHashing(t *testing.T) {

	rwc, err := tpm2.OpenTPM(tpmPath)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		data    []byte
		gostAlg tpm2.Algorithm
		shaAlg  tpm2.Algorithm
	}{
		{
			name:    "gost3411_256 + sha256",
			data:    []byte("This is test data for hashing."),
			gostAlg: tpm2.AlgGOST3411_256,
			shaAlg:  tpm2.AlgSHA256,
		},
		{
			name:    "gost3411_512 + sha512",
			data:    []byte("This is test data for hashing."),
			gostAlg: tpm2.AlgGOST3411_512,
			shaAlg:  tpm2.AlgSHA512,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gostDigets, err := hashWithTPM(rwc, testCase.data, testCase.gostAlg)
			require.NoError(t, err)

			shaDigets, err := hashWithTPM(rwc, testCase.data, testCase.shaAlg)
			require.NoError(t, err)

			require.Equal(t, gostDigets, shaDigets)
		})
	}
}

func hashWithTPM(rwc io.ReadWriteCloser, data []byte, algNum tpm2.Algorithm) ([]byte, error) {
	buf := tpmutil.U16Bytes(data)
	alg := tpm2.Algorithm(algNum)
	hierarchy := tpmutil.Handle(tpm2.HandleOwner)

	digest, _, err := tpm2.Hash(rwc, alg, buf, hierarchy)
	if err != nil {
		return nil, fmt.Errorf("Hashing failed: %v", err)
	}

	return digest, nil
}
