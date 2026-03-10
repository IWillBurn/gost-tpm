package testing

import (
	"crypto/rand"
	"fmt"
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
				Restricted:          true,
				Decrypt:             true,
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

func createSymmetricKey(
	tpm transport.TPM,
	parentHandle tpm2.TPMHandle,
	parentHandleName tpm2.TPM2BName,
	alg tpm2.TPMAlgID,
	mode tpm2.TPMAlgID,
	keyBits tpm2.TPMKeyBits,
) (tpm2.TPM2BPrivate, tpm2.TPM2BPublic, error) {
	rsp, err := tpm2.Create{
		ParentHandle: tpm2.AuthHandle{
			Handle: parentHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   parentHandleName,
		},
		InPublic: tpm2.New2B(tpm2.TPMTPublic{
			Type:    tpm2.TPMAlgSymCipher,
			NameAlg: tpm2.TPMAlgSHA256,
			ObjectAttributes: tpm2.TPMAObject{
				FixedTPM:            true,
				FixedParent:         true,
				SensitiveDataOrigin: true,
				UserWithAuth:        true,
				Decrypt:             true,
				SignEncrypt:         true,
			},
			Parameters: tpm2.NewTPMUPublicParms(tpm2.TPMAlgSymCipher, &tpm2.TPMSSymCipherParms{
				Sym: tpm2.TPMTSymDefObject{
					Algorithm: alg,
					KeyBits:   tpm2.NewTPMUSymKeyBits(alg, keyBits),
					Mode:      tpm2.NewTPMUSymMode(alg, mode),
				},
			}),
		}),
	}.Execute(tpm)
	if err != nil {
		return tpm2.TPM2BPrivate{}, tpm2.TPM2BPublic{}, err
	}
	return rsp.OutPrivate, rsp.OutPublic, nil
}

func loadKey(
	tpm transport.TPM,
	parentHandle tpm2.TPMHandle,
	parentHandleName tpm2.TPM2BName,
	priv tpm2.TPM2BPrivate,
	pub tpm2.TPM2BPublic,
) (tpm2.TPMHandle, tpm2.TPM2BName, error) {
	rsp, err := tpm2.Load{
		ParentHandle: tpm2.AuthHandle{
			Handle: parentHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   parentHandleName,
		},
		InPrivate: priv,
		InPublic:  pub,
	}.Execute(tpm)
	if err != nil {
		return 0, tpm2.TPM2BName{}, err
	}
	return rsp.ObjectHandle, rsp.Name, nil
}

func encryptSymmetric(
	tpm transport.TPM,
	keyHandle tpm2.TPMHandle,
	keyHandleName tpm2.TPM2BName,
	iv []byte,
	plaintext []byte,
) ([]byte, error) {
	rsp, err := tpm2.EncryptDecrypt2{
		KeyHandle: tpm2.AuthHandle{
			Handle: keyHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   keyHandleName,
		},
		Message: tpm2.TPM2BMaxBuffer{Buffer: plaintext},
		Decrypt: false,
		IV:      tpm2.TPM2BIV{Buffer: iv},
	}.Execute(tpm)
	if err != nil {
		return nil, err
	}
	return rsp.OutData.Buffer, nil
}

func decryptSymmetric(
	tpm transport.TPM,
	keyHandle tpm2.TPMHandle,
	keyHandleName tpm2.TPM2BName,
	iv []byte,
	ciphertext []byte,
) ([]byte, error) {
	rsp, err := tpm2.EncryptDecrypt2{
		KeyHandle: tpm2.AuthHandle{
			Handle: keyHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   keyHandleName,
		},
		Message: tpm2.TPM2BMaxBuffer{Buffer: ciphertext},
		Decrypt: true,
		IV:      tpm2.TPM2BIV{Buffer: iv},
	}.Execute(tpm)
	if err != nil {
		return nil, err
	}
	return rsp.OutData.Buffer, nil
}

func hashWithTPM(
	tpm transport.TPM,
	data []byte,
	alg tpm2.TPMAlgID,
) ([]byte, error) {
	rsp, err := tpm2.Hash{
		Data:      tpm2.TPM2BMaxBuffer{Buffer: data},
		HashAlg:   alg,
		Hierarchy: tpm2.TPMRHOwner,
	}.Execute(tpm)
	if err != nil {
		return nil, err
	}
	return rsp.OutHash.Buffer, nil
}

func hashSequenceWithChunks(
	tpm transport.TPM,
	chunks [][]byte,
	alg tpm2.TPMAlgID,
	saveRestoreContext bool,
) ([]byte, error) {
	if !saveRestoreContext {
		// Hash all data at once
		flat := make([]byte, 0)
		for _, chunk := range chunks {
			flat = append(flat, chunk...)
		}
		return hashWithTPM(tpm, flat, alg)
	}

	// Hash with sequence + context save/load between each update
	startRsp, err := tpm2.HashSequenceStart{
		HashAlg: alg,
	}.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("HashSequenceStart failed: %v", err)
	}
	seqHandle := startRsp.SequenceHandle

	var savedContext *tpm2.TPMSContext

	for _, chunk := range chunks {
		if savedContext != nil {
			loadRsp, err := tpm2.ContextLoad{
				Context: *savedContext,
			}.Execute(tpm)
			if err != nil {
				return nil, fmt.Errorf("ContextLoad failed: %v", err)
			}
			seqHandle = loadRsp.LoadedHandle
		}

		_, err = tpm2.SequenceUpdate{
			SequenceHandle: tpm2.AuthHandle{
				Handle: seqHandle,
				Auth:   tpm2.PasswordAuth(nil),
				Name:   tpm2.TPM2BName{Buffer: nil},
			},
			Buffer: tpm2.TPM2BMaxBuffer{Buffer: chunk},
		}.Execute(tpm)
		if err != nil {
			return nil, fmt.Errorf("SequenceUpdate failed: %v", err)
		}

		saveRsp, err := tpm2.ContextSave{
			SaveHandle: seqHandle,
		}.Execute(tpm)
		if err != nil {
			return nil, fmt.Errorf("ContextSave failed: %v", err)
		}
		savedContext = &saveRsp.Context

		_, _ = tpm2.FlushContext{FlushHandle: seqHandle}.Execute(tpm)
	}

	if savedContext != nil {
		loadRsp, err := tpm2.ContextLoad{
			Context: *savedContext,
		}.Execute(tpm)
		if err != nil {
			return nil, fmt.Errorf("ContextLoad failed: %v", err)
		}
		seqHandle = loadRsp.LoadedHandle
	}

	completeRsp, err := tpm2.SequenceComplete{
		SequenceHandle: tpm2.AuthHandle{
			Handle: seqHandle,
			Auth:   tpm2.PasswordAuth(nil),
			Name:   tpm2.TPM2BName{Buffer: nil},
		},
		Hierarchy: tpm2.TPMRHNull,
	}.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("SequenceComplete failed: %v", err)
	}

	return completeRsp.Result.Buffer, nil
}

func getCapabilityAlg(
	tpm transport.TPM,
	alg tpm2.TPMAlgID,
) (bool, error) {
	rsp, err := tpm2.GetCapability{
		Capability:    tpm2.TPMCapAlgs,
		Property:      uint32(alg),
		PropertyCount: 1,
	}.Execute(tpm)
	if err != nil {
		return false, err
	}

	algList, err := rsp.CapabilityData.Data.Algorithms()
	if err != nil {
		return false, fmt.Errorf("failed to extract algorithms: %v", err)
	}

	for _, a := range algList.AlgProperties {
		if a.Alg == alg {
			return true, nil
		}
	}

	return false, nil
}
