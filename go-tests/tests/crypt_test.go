package testing

import (
	"crypto/rand"
	"fmt"
	"io"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IWillBurn/gost-go-tpm/legacy/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpmutil"
)

func Test_CryptDecrypt(t *testing.T) {

	rwc, err := tpm2.OpenTPM(tpmPath)
	require.NoError(t, err)

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
		alg     tpm2.Algorithm
		mod     tpm2.Algorithm
		keySize uint16
		ivSize  int
		err     error
	}{
		{
			name:    "Test Magma-CBC",
			data:    baseData,
			alg:     tpm2.AlgMagma,
			mod:     tpm2.AlgCBC,
			keySize: 256,
			ivSize:  8,
			err:     nil,
		},
		{
			name:    "Test Magma-CTR",
			data:    baseData,
			alg:     tpm2.AlgMagma,
			mod:     tpm2.AlgCTR,
			keySize: 256,
			ivSize:  8,
			err:     nil,
		},
		{
			name:    "Test Grasshopper-CBC",
			data:    baseData,
			alg:     tpm2.AlgGrasshopper,
			mod:     tpm2.AlgCBC,
			keySize: 256,
			ivSize:  16,
			err:     nil,
		},
		{
			name:    "Test Grasshopper-CTR",
			data:    baseData,
			alg:     tpm2.AlgGrasshopper,
			mod:     tpm2.AlgCTR,
			keySize: 256,
			ivSize:  16,
			err:     nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			primary, err := CreatePrimaryKey(rwc)
			require.NoError(t, err)
			defer tpm2.FlushContext(rwc, primary)

			private, public, err := GenerateKey(rwc, testCase.alg, testCase.mod, testCase.keySize, primary)
			require.NoError(t, err)

			iv, err := generateIV(testCase.ivSize)
			require.NoError(t, err)

			enc, err := SymmetricEncryptWithTPM(rwc, private, public, iv, testCase.data, primary)
			require.NoError(t, err)

			dec, err := SymmetricDecryptWithTPM(rwc, private, public, iv, enc, primary)
			require.NoError(t, err)

			require.Equal(t, testCase.data, dec)
		})
	}
}

func CreatePrimaryKey(rwc io.ReadWriteCloser) (tpmutil.Handle, error) {
	template := tpm2.Public{
		Type:    tpm2.AlgRSA,
		NameAlg: tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
			tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt,
		AuthPolicy: nil,
		RSAParameters: &tpm2.RSAParams{
			Symmetric: &tpm2.SymScheme{
				Alg:     tpm2.AlgAES,
				KeyBits: 128,
				Mode:    tpm2.AlgCFB,
			},
			KeyBits:     2048,
			ExponentRaw: 65537,
			ModulusRaw:  make([]byte, 256),
		},
	}

	primaryHandle, _, err := tpm2.CreatePrimary(
		rwc,
		tpm2.HandleOwner,
		tpm2.PCRSelection{},
		"",
		"",
		template,
	)

	if err != nil {
		return tpm2.HandleNull, fmt.Errorf("CreatePrimary failed: %v", err)
	}

	return primaryHandle, err
}

func GenerateKey(rwc io.ReadWriteCloser, alg tpm2.Algorithm, mod tpm2.Algorithm, keySize uint16, primary tpmutil.Handle) ([]byte, []byte, error) {
	template := tpm2.Public{
		Type:    tpm2.AlgSymCipher,
		NameAlg: tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
			tpm2.FlagUserWithAuth | tpm2.FlagDecrypt | tpm2.FlagSign,
		SymCipherParameters: &tpm2.SymCipherParams{
			Symmetric: &tpm2.SymScheme{
				Alg:     alg,
				KeyBits: keySize,
				Mode:    mod,
			},
		},
	}

	private, public, _, _, _, err := tpm2.CreateKey(
		rwc,
		primary,
		tpm2.PCRSelection{},
		"",
		"",
		template,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("CreateKey failed: %v", err)
	}

	return private, public, nil
}

func SymmetricEncryptWithTPM(rwc io.ReadWriteCloser, private []byte, public []byte, iv []byte, in []byte, primary tpmutil.Handle) ([]byte, error) {
	keyHandle, _, err := tpm2.Load(rwc, primary, "", public, private)
	if err != nil {
		return nil, fmt.Errorf("Load failed: %v", err)
	}
	defer tpm2.FlushContext(rwc, keyHandle)

	out, err := tpm2.EncryptSymmetric(rwc, "", keyHandle, iv, in)
	if err != nil {
		return nil, fmt.Errorf("EncryptSymmetric failed: %v", err)
	}

	return out, nil
}

func SymmetricDecryptWithTPM(rwc io.ReadWriteCloser, private []byte, public []byte, iv []byte, in []byte, primary tpmutil.Handle) ([]byte, error) {
	keyHandle, _, err := tpm2.Load(rwc, primary, "", public, private)
	if err != nil {
		return nil, fmt.Errorf("Load failed: %v", err)
	}
	defer tpm2.FlushContext(rwc, keyHandle)

	out, err := tpm2.DecryptSymmetric(rwc, "", keyHandle, iv, in)
	if err != nil {
		return nil, fmt.Errorf("DecryptSymmetric failed: %v", err)
	}

	return out, nil
}

func generateIV(size int) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}

	iv := make([]byte, size)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}
	return iv, nil
}
