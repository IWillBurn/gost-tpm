package testing

import (
	"testing"

	"github.com/IWillBurn/gost-go-tpm/tpm2"
	"github.com/IWillBurn/gost-go-tpm/tpm2/transport"
	"github.com/stretchr/testify/require"
)

func flush(tpm transport.TPM, h tpm2.TPMHandle) {
	_, _ = tpm2.FlushContext{FlushHandle: h}.Execute(tpm)
}

func mustCreatePrimary(
	t *testing.T,
	tpm transport.TPM,
) (tpm2.TPMHandle, tpm2.TPM2BName) {
	t.Helper()

	rsp, err := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic: tpm2.New2B(tpm2.TPMTPublic{
			Type:    tpm2.TPMAlgRSA,
			NameAlg: tpm2.TPMAlgSHA256,
			ObjectAttributes: tpm2.TPMAObject{
				FixedTPM:            true,
				FixedParent:         true,
				SensitiveDataOrigin: true,
				UserWithAuth:        true,

				Restricted: true,
				Decrypt:    true,
			},
			Parameters: tpm2.NewTPMUPublicParms(tpm2.TPMAlgRSA, &tpm2.TPMSRSAParms{
				Symmetric: tpm2.TPMTSymDefObject{
					Algorithm: tpm2.TPMAlgAES,
					KeyBits:   tpm2.NewTPMUSymKeyBits(tpm2.TPMAlgAES, tpm2.TPMKeyBits(128)),
					Mode:      tpm2.NewTPMUSymMode(tpm2.TPMAlgAES, tpm2.TPMAlgCFB),
				},
				Scheme: tpm2.TPMTRSAScheme{
					Scheme: tpm2.TPMAlgNull,
				},
				KeyBits:  2048,
				Exponent: 65537,
			}),
			Unique: tpm2.NewTPMUPublicID(tpm2.TPMAlgRSA, &tpm2.TPM2BPublicKeyRSA{
				Buffer: nil,
			}),
		}),
	}.Execute(tpm)
	require.NoError(t, err)

	return rsp.ObjectHandle, rsp.Name
}
