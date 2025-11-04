package testing

import (
	"bytes"
	"fmt"
	"io"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IWillBurn/gost-go-tpm/legacy/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpmutil"
)

func Test_NVMarshaling_Hashing(t *testing.T) {

	rwc, err := tpm2.OpenTPM(tpmPath)
	require.NoError(t, err)

	testCases := []struct {
		name string
		data [][]byte
		alg  tpm2.Algorithm
	}{
		{
			name: "AlgGOST3411_256",
			data: [][]byte{
				[]byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37},
				[]byte{0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35},
				[]byte{0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33},
				[]byte{0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31},
				[]byte{0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39},
				[]byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37},
				[]byte{0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35},
				[]byte{0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32},
			},
			alg: tpm2.AlgGOST3411_256,
		},
		{
			name: "AlgGOST3411_512",
			data: [][]byte{
				[]byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37},
				[]byte{0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35},
				[]byte{0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33},
				[]byte{0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31},
				[]byte{0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39},
				[]byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37},
				[]byte{0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35},
				[]byte{0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32},
			},
			alg: tpm2.AlgGOST3411_512,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			digestWithoutInterruptions, err := hashSequenceWithoutInterruptions(rwc, testCase.data, testCase.alg)
			require.NoError(t, err)

			digestWithInterruptions, err := hashSequenceWithInterruptions(rwc, testCase.data, testCase.alg)
			require.NoError(t, err)

			require.Equal(t, digestWithoutInterruptions, digestWithInterruptions)
		})
	}
}

func hashSequenceWithoutInterruptions(rwc io.ReadWriteCloser, data [][]byte, alg tpm2.Algorithm) ([]byte, error) {
	flat := bytes.Join(data, []byte{})

	buf := tpmutil.U16Bytes(flat)
	hierarchy := tpmutil.Handle(tpm2.HandleOwner)

	digest, _, err := tpm2.Hash(rwc, alg, buf, hierarchy)
	if err != nil {
		return nil, fmt.Errorf("Hashing failed: %v", err)
	}

	return digest, nil
}

func hashSequenceWithInterruptions(rwc io.ReadWriteCloser, data [][]byte, alg tpm2.Algorithm) ([]byte, error) {
	sequenceHandle, err := tpm2.HashSequenceStart(rwc, "", alg)
	if err != nil {
		return nil, fmt.Errorf("HashSequenceStart failed: %v", err)
	}
	defer tpm2.FlushContext(rwc, sequenceHandle)

	var sevedContext []byte

	for _, d := range data {
		if sevedContext != nil {
			sequenceHandle, err = tpm2.ContextLoad(rwc, sevedContext)
			if err != nil {
				return nil, fmt.Errorf("ContextLoad failed: %v", err)
			}
		}

		buf := tpmutil.U16Bytes(d)
		err = tpm2.SequenceUpdate(rwc, "", sequenceHandle, buf)
		if err != nil {
			return nil, fmt.Errorf("SequenceUpdate failed: %v", err)
		}

		sevedContext, err = tpm2.ContextSave(rwc, sequenceHandle)
		if err != nil {
			return nil, fmt.Errorf("ContextSave failed: %v", err)
		}

		err = tpm2.FlushContext(rwc, sequenceHandle)
		if err != nil {
			return nil, fmt.Errorf("FlushContext failed: %v", err)
		}
	}

	if sevedContext != nil {
		sequenceHandle, err = tpm2.ContextLoad(rwc, sevedContext)
		if err != nil {
			return nil, fmt.Errorf("ContextLoad failed: %v", err)
		}
	}

	digest, _, err := tpm2.SequenceComplete(rwc, "", sequenceHandle, tpm2.HandleNull, nil)
	if err != nil {
		return nil, fmt.Errorf("SequenceComplete failed: %v", err)
	}

	return digest, nil
}
