# CHANGELOG

---

## GOST-LIBTPMS

**Repository:** [github.com/IWillBurn/gost-libtpms](https://github.com/IWillBurn/gost-libtpms)

### [1] Added support for GOST R 34.11-2012 (Streebog) hashing

**Algorithms:** `GOST3411-256` (ALG ID `0x0100`), `GOST3411-512` (ALG ID `0x0101`)

- Registered new hash algorithms `TPM_ALG_GOST3411_256` and `TPM_ALG_GOST3411_512` in the TPM algorithm table (`TpmTypes.h`, `AlgorithmCap.c`).
- Integrated hash implementation from the gost-engine library (`gosthash2012.h`): functions `init_gost2012_hash_ctx`, `gost2012_hash_block`, `gost2012_finish_hash` are connected through wrappers `GOST3411_256_INIT/UPDATE/FINAL` and `GOST3411_512_INIT/UPDATE/FINAL` (`TpmToOsslHash.h`).
- Updated reference hash values for self-test (`HashTestData.h`) — now contain correct GOST R 34.11-2012 digests.
- Implemented marshaling/unmarshaling functions for hash context state (`NVMarshal.c`): fields `buffer`, `h`, `N`, `Sigma`, `bufsize`, `digest_size` of the `gost2012_hash_ctx` structure are serialized for correct context save/restore (NV storage, context save/load).
- Algorithms added to the `custom:gost` runtime profile (`RuntimeProfile.c`, `RuntimeAlgorithm.c`).

### [2] Added support for Magma and Grasshopper (Kuznechik) symmetric encryption

**Algorithms:** `Magma` (ALG ID `0x0102`), `Grasshopper` (ALG ID `0x0103`)

- Registered algorithms `TPM_ALG_MAGMA` and `TPM_ALG_GRASSHOPPER` as symmetric block ciphers (`TpmTypes.h`, `TpmProfile_Common.h`, `AlgorithmCap.c`).
- Defined key and block parameters:
  - **Magma:** 256-bit key, 8-byte block; CTR, CBC modes.
  - **Grasshopper:** 256-bit key, 16-byte block; CTR, OFB, CBC, CFB, ECB modes.
- Created module `TpmToGostSupport.c` / `TpmToGostSupport_fp.h` — wrappers over gost-engine functions (`magma_key`, `magmacrypt`, `magmadecrypt`, `grasshopper_set_encrypt_key`, `grasshopper_encrypt_block`, etc.) for integration into the TPM CryptSym infrastructure.
- Defined macros `TpmCryptSetEncryptKey*`, `TpmCryptSetDecryptKey*`, `TpmCryptEncrypt*`, `TpmCryptDecrypt*` for both algorithms (`TpmToOsslSym.h`).
- Added test vectors from GOST R 34.13-2015 (`SymmetricTestData.h`, `SymmetricTest.h`): CTR and CBC for Magma, CTR and CBC for Grasshopper.
- Updated key and mode marshaling/unmarshaling functions (`Marshal.c`, `Unmarshal.c`): added handling for `TPMI_MAGMA_KEY_BITS`, `TPMI_GRASSHOPPER_KEY_BITS`, `TPMU_SYM_KEY_BITS`, `TPMU_SYM_MODE`.
- Updated EVP wrappers for OpenSSL (`Helpers.c`): mapping `TPM_ALG_MAGMA` / `TPM_ALG_GRASSHOPPER` → corresponding `EVP_CIPHER` from gost-engine.
- Algorithms included in self-test (`AlgorithmTests.c`), `FOR_EACH_SYM` macro (`CryptSym.h`), runtime profile.
- Updated `MAX_SYM_BLOCK_SIZE` and `MAX_SYM_KEY_BITS` to account for the new algorithms (`TpmAlgorithmDefines.h`).

### [3] Added support for GOST R 34.10-2012 digital signatures

**Algorithms:** `GOST3410-256` (ALG ID `0x0104`), `GOST3410-512` (ALG ID `0x0105`)

- Registered signature algorithms `TPM_ALG_GOST3410_256` and `TPM_ALG_GOST3410_512` (`TpmTypes.h`, `AlgorithmCap.c`).
- Created module `TpmEcc_Signature_GOST3410.c` / `TpmEcc_Signature_GOST3410_fp.h` with two implementations:
  - **Pure TPM math** (default): sign/verify implementation per GOST R 34.10-2012 using internal TPM bignum operations (`ExtMath_*`, `TpmEcc_PointMult`, `TpmEcc_GenerateKeyPair`). Includes correct digest handling in little-endian format (`TpmEcc_AdjustGost3410Digest`).
  - **Via gost-engine** (optional, `USE_OPENSSL_FUNCTIONS_GOST3410`): delegates signing/verification to `gost_ec_sign` / `gost_ec_verify`.
- Added 7 TC26 elliptic curves (`TpmTypes.h`, `TpmAlgorithmDefines.h`, `EccConstantData.inl`, `BnEccConstants.c`, `CryptEccData.c`, `OIDs.h`):
  - **256-bit:** `tc26-gost-3410-2012-256-paramSetA/B/C/D` (curve IDs `0x0006`–`0x0009`)
  - **512-bit:** `tc26-gost-3410-2012-512-paramSetA/B/C` (curve IDs `0x000A`–`0x000C`)
  - Parameters p, a, b, n, h, G(x,y) defined in `ECC_CONST` format for each curve.
  - OIDs defined per RFC 7836 / TC26.
- Integrated into the signature dispatcher (`CryptEccSignature.c`): `CryptEccSign` and `CryptEccValidateSignature` route calls for `TPM_ALG_GOST3410_256` / `512`.
- Added types `TPMS_SIG_SCHEME_GOST3410_256/512`, `TPMS_SIGNATURE_GOST3410_256/512` and corresponding union members in `TPMU_SIG_SCHEME`, `TPMU_ASYM_SCHEME`, `TPMU_SIGNATURE` (`TpmTypes.h`).
- Implemented marshaling/unmarshaling functions for all new types (`Marshal.c`, `Unmarshal.c`, `Marshal_fp.h`, `Unmarshal_fp.h`).
- Added test data for self-test (`EccTestData.h`): test keys (d, Qx, Qy) and signature values for 256-bit and 512-bit curves.
- Self-test (`AlgorithmTests.c`): added `TestGOST3410SignAndVerify` function that performs sign+verify during TPM initialization.
- Updated `CryptUtil.c`, `CryptIsAsymSignScheme`, `CryptGetSignHashAlg` to recognize GOST signature schemes.
- Curves registered in runtime algorithms with shortcut `ecc-tc26-gost3410` (`RuntimeAlgorithm.c`).
- `TPM2B_EC_TEST` size increased from 32 to 64 bytes to support 512-bit curves.

### [4] Build configuration and TPM 1.2 updates

- Added linking with `libgost` in `libtpms.pc.in`.
- Added `configure.ac`: option `use_openssl_functions_gost3410`.
- Updated `Makefile.am`: added new source files (`TpmEcc_Signature_GOST3410.c`, `TpmToGostSupport.c`) and headers.
- TPM 1.2 (tpm12): added identifiers `TPM_ALG_MAGMA256` (`0x0B`) and `TPM_ALG_GRASSHOPPER256` (`0x0C`) in `tpm_constants.h`, updated `tpm_key.c`, `tpm_process.c`, `tpm_transport.c` to recognize the new algorithms.
- Updated compile-constants table in `NVMarshal.c` (array size 122→133) to support new algorithms and curves during state migration.

---

## GOST-GO-TPM

**Repository:** [github.com/IWillBurn/gost-go-tpm](https://github.com/IWillBurn/gost-go-tpm)


### [1] Added GOST algorithm constants

**Files:** `tpm2/constants.go`, `legacy/tpm2/constants.go`, `tpm/constants.go`

**TPM 2.0 API (`tpm2/constants.go`):**

- `TPMAlgGOST3411256` = `0x0100`, `TPMAlgGOST3411512` = `0x0101` (hashing)
- `TPMAlgMagma` = `0x0102`, `TPMAlgGrasshopper` = `0x0103` (symmetric encryption)
- `TPMAlgGOST3410256` = `0x0104`, `TPMAlgGOST3410512` = `0x0105` (digital signatures)
- TC26 curves: `TPMECCCurveTC26Gost3410256ParamSetA/B/C/D` (`0x0006`–`0x0009`), `TPMECCCurveTC26Gost3410512ParamSetA/B/C` (`0x000A`–`0x000C`)

**Legacy API (`legacy/tpm2/constants.go`):**

- Analogous constants: `AlgMagma`, `AlgGrasshopper`, `AlgGOST3410_256`, `AlgGOST3410_512`
- Curves: `CurveTC26Gost3410256ParamSetA/B/C/D`, `CurveTC26Gost3410512ParamSetA/B/C`
- Fixed values of `CurveBNP256` (`0x0010`) and `CurveBNP638` (`0x0011`) — previously used iota, which conflicted with the new curves.
- Added string representations in `Algorithm.String()`.

**TPM 1.2 API (`tpm/constants.go`):**

- `AlgMagma256`, `AlgGrasshopper256` with mapping in `AlgMap`.


### [2] Added GOST type support in marshaling/unmarshaling

**File:** `tpm2/structures.go`

**Symmetric algorithms:**

- `TPMUSymKeyBits`: added handling for `TPMAlgMagma` and `TPMAlgGrasshopper` in `create()`, `get()`, `Sym()`. Added methods `Magma()` and `Grasshopper()`.
- `TPMUSymMode`: analogous changes — `create()`, `get()`, `Sym()`, methods `Magma()`, `Grasshopper()`.
- `TPMUSymDetails`: added handling for `TPMAlgMagma` / `TPMAlgGrasshopper` as `TPMSEmpty`.

**Signature schemes:**

- New type `TPMSSigSchemeGOST3410` (alias for `TPMSSchemeHash`).
- `TPMUSigScheme`: added handling for `TPMAlgGOST3410256` / `TPMAlgGOST3410512` in `create()`, `get()`. Methods `GOST3410256()`, `GOST3410512()`.
- `TPMUAsymScheme`: added `TPMSSigSchemeGOST3410` to type constraint `AsymSchemeContents`. Handling in `create()`, `get()`. Methods `GOST3410256()`, `GOST3410512()`.
- `TPMUSignature`: added handling for `TPMAlgGOST3410256` / `TPMAlgGOST3410512` as `TPMSSignatureECC`. Methods `GOST3410256()`, `GOST3410512()`.

**Legacy API (`legacy/tpm2/structures.go`):**

- `DecodeSignature`: added handling for `AlgGOST3410_256` / `AlgGOST3410_512` — ECC signature decoding (R, S).

---

## GOST-TPM
**Repository:** [github.com/IWillBurn/gost-tpm](https://github.com/IWillBurn/gost-tpm)


### [1] Project initialization and submodule structure

- Created `.gitmodules` with 4 submodules:
  - **gost-libtpms** — libtpms fork with GOST algorithms
  - **gost-swtpm** — swtpm fork
  - **gost-go-tpm** — go-tpm fork with GOST support
  - **gost-engine** — gost-engine (GOST cryptographic library for OpenSSL)
- Created `.gitignore` (IDE files).
- Created `README.md` with build and run instructions.

### [2] Docker infrastructure

**Files:** `.docker/build.Dockerfile`, `.docker/tests.Dockerfile`, `.docker/scripts/`

**`build.Dockerfile` — multi-stage build:**

- **Stage 1** (`build-swtpm`): building gost-engine → gost-libtpms → gost-swtpm on Ubuntu 22.04 with a full set of dependencies.
- **Stage 2** (`runtime`): minimal image with installed libraries, swtpm, tpm2-tools.
- Script `run_build.sh`: swtpm initialization (`swtpm_setup --profile name=custom:gost`), swtpm socket startup, selftest.

**`tests.Dockerfile` — extended build for testing:**

- **Stage 1**: same as build.
- **Stage 2** (`build-tests`): building Go tests (`golang:1.25-alpine`, `go test -c`).
- **Stage 3** (`runtime`): swtpm startup + Go test execution.
- Script `run_tests.sh`: swtpm initialization, test binary execution.

### [3] Integration tests (Go)

**Directory:** `go-tests/`

- **Go module** (`go.mod`): dependency on `github.com/IWillBurn/gost-go-tpm`, `testify`.
- **Makefile** (`scripts/Makefile`): test binary compilation for `linux/amd64`.

**Utilities (`tests/crypto_utils.go`):**

- `mustCreatePrimary` — RSA primary key creation (owner hierarchy).
- `createSymmetricKey`, `loadKey`, `encryptSymmetric`, `decryptSymmetric` — symmetric key operations via `EncryptDecrypt2`.
- `hashWithTPM`, `hashSequenceWithChunks` — hashing (single and sequential with context save/restore).
- `getCapabilityAlg` — algorithm availability check via `GetCapability`.
- GOST signature functions: `createGOST3410SigningKey`, `loadGOST3410SigningKey`, `signGOST3410`, `verifySignatureByHandle`.

**Tests:**

- **`Test_HasAlgs`** (`has_algs_test.go`): verifies that the TPM reports support for all 6 GOST algorithms: `GOST3411-256`, `GOST3411-512`, `Magma`, `Grasshopper`, `GOST3410-256`, `GOST3410-512`.

- **`Test_Hash`** (`hash_test.go`): verifies correctness of GOST R 34.11-2012 hashing against reference vectors from RFC 6986 (4 test cases: M1/M2 for 256 and 512 bits).

- **`Test_NVMarshaling_Hashing`** (`marshaling_test.go`): verifies correctness of hash context marshaling: hashing data in chunks with `HashSequenceStart` → `SequenceUpdate` + `ContextSave` / `ContextLoad` → `SequenceComplete` must produce the same result as hashing in one pass. Tested for `GOST3411-256` and `GOST3411-512`.

- **`Test_CryptDecrypt`** (`crypt_test.go`): verifies symmetric encryption/decryption via `EncryptDecrypt2`. 4 test cases:
  - Magma-CBC, Magma-CTR
  - Grasshopper-CBC, Grasshopper-CTR

  A symmetric key is created, data is encrypted, then decrypted — the result is verified to match the original.

- **`Test_ECDSASign_Verify`** (`sign_test.go`): verifies GOST R 34.10-2012 signing and verification. 7 test cases across all supported curves:
  - `gost3410-256` + paramSetA/B/C/D
  - `gost3410-512` + paramSetA/B/C

  An ECC key with a GOST scheme is created, a digest is signed, and the signature is verified.