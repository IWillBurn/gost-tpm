package testing

import (
	"testing"

	"github.com/IWillBurn/gost-go-tpm/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpm2/transport"
	"github.com/stretchr/testify/require"
)

func Test_CryptDecrypt(t *testing.T) {
	tpm, err := transport.OpenTPM(tpmPath)
	require.NoError(t, err)
	defer tpm.Close()

	primary, primaryName := mustCreatePrimary(t, tpm)
	defer flush(tpm, primary)

	baseData := []byte{
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
		0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35,
		0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33,
		0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31,
		0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
	}

	testCases := []struct {
		name    string
		data    []byte
		alg     tpm2.TPMAlgID
		mode    tpm2.TPMAlgID
		keyBits tpm2.TPMKeyBits
		ivSize  int
	}{
		{
			name:    "Test Magma-CBC",
			data:    baseData,
			alg:     tpm2.TPMAlgMagma,
			mode:    tpm2.TPMAlgCBC,
			keyBits: 256,
			ivSize:  8,
		},
		{
			name:    "Test Magma-CTR",
			data:    baseData,
			alg:     tpm2.TPMAlgMagma,
			mode:    tpm2.TPMAlgCTR,
			keyBits: 256,
			ivSize:  8,
		},
		{
			name:    "Test Grasshopper-CBC",
			data:    baseData,
			alg:     tpm2.TPMAlgGrasshopper,
			mode:    tpm2.TPMAlgCBC,
			keyBits: 256,
			ivSize:  16,
		},
		{
			name:    "Test Grasshopper-CTR",
			data:    baseData,
			alg:     tpm2.TPMAlgGrasshopper,
			mode:    tpm2.TPMAlgCTR,
			keyBits: 256,
			ivSize:  16,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			priv, pub, err := createSymmetricKey(tpm, primary, primaryName, testCase.alg, testCase.mode, testCase.keyBits)
			require.NoError(t, err)

			iv, err := generateIV(testCase.ivSize)
			require.NoError(t, err)

			keyHandle, keyHandleName, err := loadKey(tpm, primary, primaryName, priv, pub)
			require.NoError(t, err)
			defer flush(tpm, keyHandle)

			encrypted, err := encryptSymmetric(tpm, keyHandle, keyHandleName, iv, testCase.data)
			require.NoError(t, err)

			// Reload key for decryption (fresh state)
			keyHandle2, keyHandleName2, err := loadKey(tpm, primary, primaryName, priv, pub)
			require.NoError(t, err)
			defer flush(tpm, keyHandle2)

			decrypted, err := decryptSymmetric(tpm, keyHandle2, keyHandleName2, iv, encrypted)
			require.NoError(t, err)

			require.Equal(t, testCase.data, decrypted)
		})
	}
}
