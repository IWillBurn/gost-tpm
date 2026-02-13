package testing

import (
	"testing"

	"github.com/IWillBurn/gost-go-tpm/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpm2/transport"
	"github.com/stretchr/testify/require"
)

func Test_ECDSASign_Verify(t *testing.T) {
	tpm, err := transport.OpenTPM(tpmPath)
	require.NoError(t, err)
	defer tpm.Close()

	primary, primaryName := mustCreatePrimary(t, tpm)
	defer flush(tpm, primary)

	digets := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}

	testCases := []struct {
		name  string
		alg   tpm2.TPMAlgID
		curve tpm2.TPMECCCurve
		hash  tpm2.TPMAlgID
		data  []byte
		err   bool
	}{
		{
			name:  "Test ECC: gost3410-256 + tc26-gost3410-256-param-set-a",
			alg:   tpm2.TPMAlgGOST3410256,
			curve: tpm2.TPMECCCurveTC26Gost3410256ParamSetA,
			hash:  tpm2.TPMAlgGOST3411256,
			data:  digets[:32],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-256 + tc26-gost3410-256-param-set-b",
			alg:   tpm2.TPMAlgGOST3410256,
			curve: tpm2.TPMECCCurveTC26Gost3410256ParamSetB,
			hash:  tpm2.TPMAlgGOST3411256,
			data:  digets[:32],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-256 + tc26-gost3410-256-param-set-c",
			alg:   tpm2.TPMAlgGOST3410256,
			curve: tpm2.TPMECCCurveTC26Gost3410256ParamSetC,
			hash:  tpm2.TPMAlgGOST3411256,
			data:  digets[:32],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-256 + tc26-gost3410-256-param-set-d",
			alg:   tpm2.TPMAlgGOST3410256,
			curve: tpm2.TPMECCCurveTC26Gost3410256ParamSetD,
			hash:  tpm2.TPMAlgGOST3411256,
			data:  digets[:32],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-512 + tc26-gost3410-512-param-set-a",
			alg:   tpm2.TPMAlgGOST3410512,
			curve: tpm2.TPMECCCurveTC26Gost3410512ParamSetA,
			hash:  tpm2.TPMAlgGOST3411512,
			data:  digets[:64],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-512 + tc26-gost3410-512-param-set-b",
			alg:   tpm2.TPMAlgGOST3410512,
			curve: tpm2.TPMECCCurveTC26Gost3410512ParamSetB,
			hash:  tpm2.TPMAlgGOST3411512,
			data:  digets[:64],
			err:   false,
		},
		{
			name:  "Test ECC: gost3410-512 + tc26-gost3410-512-param-set-c",
			alg:   tpm2.TPMAlgGOST3410512,
			curve: tpm2.TPMECCCurveTC26Gost3410512ParamSetC,
			hash:  tpm2.TPMAlgGOST3411512,
			data:  digets[:64],
			err:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			priv, pub, err := createGOST3410SigningKey(tpm, primary, primaryName, testCase.alg, testCase.curve, testCase.hash)
			require.NoError(t, err)

			keyHandle, keyHandleName, err := loadGOST3410SigningKey(tpm, primary, primaryName, priv, pub)
			require.NoError(t, err)
			defer flush(tpm, keyHandle)

			signature, err := signGOST3410(tpm, keyHandle, keyHandleName, testCase.alg, testCase.hash, testCase.data)
			require.NoError(t, err)

			err = verifySignatureByHandle(tpm, keyHandle, keyHandleName, testCase.data, signature)
			require.NoError(t, err)
		})
	}
}

func createGOST3410SigningKey(
	tpm transport.TPM,
	parentHanle tpm2.TPMHandle,
	parentHandleName tpm2.TPM2BName,
	alg tpm2.TPMAlgID,
	curve tpm2.TPMECCCurve,
	hash tpm2.TPMAlgID,
) (tpm2.TPM2BPrivate, tpm2.TPM2BPublic, error) {
	rsp, err := tpm2.Create{
		ParentHandle: tpm2.AuthHandle{
			Handle: parentHanle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   parentHandleName,
		},
		InPublic: tpm2.New2B(tpm2.TPMTPublic{
			Type:    tpm2.TPMAlgECC,
			NameAlg: hash,
			ObjectAttributes: tpm2.TPMAObject{
				FixedTPM:            true,
				FixedParent:         true,
				SensitiveDataOrigin: true,
				UserWithAuth:        true,
				SignEncrypt:         true,
			},
			Parameters: tpm2.NewTPMUPublicParms(tpm2.TPMAlgECC, &tpm2.TPMSECCParms{
				CurveID: curve,
				Scheme: tpm2.TPMTECCScheme{
					Scheme: alg,
					Details: tpm2.NewTPMUAsymScheme(alg, &tpm2.TPMSSigSchemeGOST3410{
						HashAlg: hash,
					}),
				},
				KDF: tpm2.TPMTKDFScheme{Scheme: tpm2.TPMAlgNull},
			}),
		}),
	}.Execute(tpm)

	return rsp.OutPrivate, rsp.OutPublic, err
}

func loadGOST3410SigningKey(
	tpm transport.TPM,
	parentHanle tpm2.TPMHandle,
	parentHandleName tpm2.TPM2BName,
	priv tpm2.TPM2BPrivate,
	pub tpm2.TPM2BPublic,
) (tpm2.TPMHandle, tpm2.TPM2BName, error) {

	rsp, err := tpm2.Load{
		ParentHandle: tpm2.AuthHandle{
			Handle: parentHanle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   parentHandleName,
		},
		InPrivate: priv,
		InPublic:  pub,
	}.Execute(tpm)

	return rsp.ObjectHandle, rsp.Name, err
}

func signGOST3410(
	tpm transport.TPM,
	keyHandle tpm2.TPMHandle,
	keyHandleName tpm2.TPM2BName,
	alg tpm2.TPMAlgID,
	hash tpm2.TPMAlgID,
	digest []byte,
) (tpm2.TPMTSignature, error) {

	rsp, err := tpm2.Sign{
		KeyHandle: tpm2.AuthHandle{
			Handle: keyHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   keyHandleName,
		},
		Digest: tpm2.TPM2BDigest{Buffer: digest},
		InScheme: tpm2.TPMTSigScheme{
			Scheme: alg,
			Details: tpm2.NewTPMUSigScheme(alg, &tpm2.TPMSSchemeHash{
				HashAlg: hash,
			}),
		},
		Validation: tpm2.TPMTTKHashCheck{
			Tag:       tpm2.TPMSTHashCheck,
			Hierarchy: tpm2.TPMRHNull,
		},
	}.Execute(tpm)

	return rsp.Signature, err
}

func verifySignatureByHandle(
	tpm transport.TPM,
	keyHandle tpm2.TPMHandle,
	keyHandleName tpm2.TPM2BName,
	digest []byte,
	sig tpm2.TPMTSignature,
) error {

	_, err := tpm2.VerifySignature{
		KeyHandle: tpm2.NamedHandle{Handle: keyHandle, Name: keyHandleName},
		Digest:    tpm2.TPM2BDigest{Buffer: digest},
		Signature: sig,
	}.Execute(tpm)

	return err
}
