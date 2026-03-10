# CHANGELOG

---

## GOST-LIBTPMS

**Репозиторий:** [github.com/IWillBurn/gost-libtpms](https://github.com/IWillBurn/gost-libtpms)

### [1] Добавлена поддержка хеширования ГОСТ Р 34.11-2012 (Стрибог)

**Алгоритмы:** `GOST3411-256` (ALG ID `0x0100`), `GOST3411-512` (ALG ID `0x0101`)

- Зарегистрированы новые алгоритмы хеширования `TPM_ALG_GOST3411_256` и `TPM_ALG_GOST3411_512` в таблице алгоритмов TPM (`TpmTypes.h`, `AlgorithmCap.c`).
- Интегрирована реализация хеширования из библиотеки gost-engine (`gosthash2012.h`): функции `init_gost2012_hash_ctx`, `gost2012_hash_block`, `gost2012_finish_hash` подключены через обёртки `GOST3411_256_INIT/UPDATE/FINAL` и `GOST3411_512_INIT/UPDATE/FINAL` (`TpmToOsslHash.h`).
- Обновлены эталонные хеш-значения для self-test (`HashTestData.h`) — содержат корректные дайджесты ГОСТ Р 34.11-2012.
- Реализованы функции маршалинга/демаршалинга состояния хеш-контекста (`NVMarshal.c`): сериализуются поля `buffer`, `h`, `N`, `Sigma`, `bufsize`, `digest_size` структуры `gost2012_hash_ctx` для корректного сохранения/восстановления контекста (NV-хранилище, context save/load).
- Алгоритмы добавлены в runtime-профиль `custom:gost` (`RuntimeProfile.c`, `RuntimeAlgorithm.c`).

### [2] Добавлена поддержка симметричного шифрования Магма и Кузнечик

**Алгоритмы:** `Magma` (ALG ID `0x0102`), `Grasshopper` (ALG ID `0x0103`)

- Зарегистрированы алгоритмы `TPM_ALG_MAGMA` и `TPM_ALG_GRASSHOPPER` как симметричные блочные шифры (`TpmTypes.h`, `TpmProfile_Common.h`, `AlgorithmCap.c`).
- Определены параметры ключей и блоков:
  - **Магма:** ключ 256 бит, блок 8 байт; режимы CTR, CBC.
  - **Кузнечик:** ключ 256 бит, блок 16 байт; режимы CTR, OFB, CBC, CFB, ECB.
- Создан модуль `TpmToGostSupport.c` / `TpmToGostSupport_fp.h` — обёртки над функциями gost-engine (`magma_key`, `magmacrypt`, `magmadecrypt`, `grasshopper_set_encrypt_key`, `grasshopper_encrypt_block` и др.) для интеграции в инфраструктуру CryptSym TPM.
- Определены макросы `TpmCryptSetEncryptKey*`, `TpmCryptSetDecryptKey*`, `TpmCryptEncrypt*`, `TpmCryptDecrypt*` для обоих алгоритмов (`TpmToOsslSym.h`).
- Добавлены тестовые векторы из ГОСТ Р 34.13-2015 (`SymmetricTestData.h`, `SymmetricTest.h`): для Магмы — CTR и CBC, для Кузнечика — CTR и CBC.
- Обновлены функции маршалинга/демаршалинга ключей и режимов (`Marshal.c`, `Unmarshal.c`): добавлена обработка `TPMI_MAGMA_KEY_BITS`, `TPMI_GRASSHOPPER_KEY_BITS`, `TPMU_SYM_KEY_BITS`, `TPMU_SYM_MODE`.
- Обновлены EVP-обёртки для OpenSSL (`Helpers.c`): маппинг `TPM_ALG_MAGMA` / `TPM_ALG_GRASSHOPPER` → соответствующие `EVP_CIPHER` из gost-engine.
- Алгоритмы включены в self-test (`AlgorithmTests.c`), `FOR_EACH_SYM` макрос (`CryptSym.h`), runtime-профиль.
- Обновлены `MAX_SYM_BLOCK_SIZE` и `MAX_SYM_KEY_BITS` с учётом новых алгоритмов (`TpmAlgorithmDefines.h`).

### [3] Добавлена поддержка ЭЦП ГОСТ Р 34.10-2012

**Алгоритмы:** `GOST3410-256` (ALG ID `0x0104`), `GOST3410-512` (ALG ID `0x0105`)

- Зарегистрированы алгоритмы подписи `TPM_ALG_GOST3410_256` и `TPM_ALG_GOST3410_512` (`TpmTypes.h`, `AlgorithmCap.c`).
- Создан модуль `TpmEcc_Signature_GOST3410.c` / `TpmEcc_Signature_GOST3410_fp.h` с двумя реализациями:
  - **Чистая TPM-математика** (по умолчанию): реализация sign/verify по ГОСТ Р 34.10-2012 через внутренние bignum-операции TPM (`ExtMath_*`, `TpmEcc_PointMult`, `TpmEcc_GenerateKeyPair`). Включает корректную обработку дайджеста в little-endian формате (`TpmEcc_AdjustGost3410Digest`).
  - **Через gost-engine** (опциональная, `USE_OPENSSL_FUNCTIONS_GOST3410`): делегирует подпись/верификацию в `gost_ec_sign` / `gost_ec_verify`.
- Добавлены 7 эллиптических кривых ТК26 (`TpmTypes.h`, `TpmAlgorithmDefines.h`, `EccConstantData.inl`, `BnEccConstants.c`, `CryptEccData.c`, `OIDs.h`):
  - **256-бит:** `tc26-gost-3410-2012-256-paramSetA/B/C/D` (curve IDs `0x0006`–`0x0009`)
  - **512-бит:** `tc26-gost-3410-2012-512-paramSetA/B/C` (curve IDs `0x000A`–`0x000C`)
  - Для каждой кривой определены параметры p, a, b, n, h, G(x,y) в формате `ECC_CONST`.
  - Определены OID по RFC 7836 / ТК26.
- Интегрированы в диспетчер подписей (`CryptEccSignature.c`): `CryptEccSign` и `CryptEccValidateSignature` маршрутизируют вызовы для `TPM_ALG_GOST3410_256` / `512`.
- Добавлены типы `TPMS_SIG_SCHEME_GOST3410_256/512`, `TPMS_SIGNATURE_GOST3410_256/512` и соответствующие union-члены в `TPMU_SIG_SCHEME`, `TPMU_ASYM_SCHEME`, `TPMU_SIGNATURE` (`TpmTypes.h`).
- Реализованы функции маршалинга/демаршалинга для всех новых типов (`Marshal.c`, `Unmarshal.c`, `Marshal_fp.h`, `Unmarshal_fp.h`).
- Добавлены тестовые данные для self-test (`EccTestData.h`): тестовые ключи (d, Qx, Qy) и значения для подписи для кривых 256 и 512 бит.
- Self-test (`AlgorithmTests.c`): добавлена функция `TestGOST3410SignAndVerify`, выполняющая sign+verify при инициализации TPM.
- Обновлён `CryptUtil.c`, `CryptIsAsymSignScheme`, `CryptGetSignHashAlg` для распознавания ГОСТ-схем подписи.
- Кривые зарегистрированы в runtime-алгоритмах с shortcut `ecc-tc26-gost3410` (`RuntimeAlgorithm.c`).
- Размер `TPM2B_EC_TEST` увеличен с 32 до 64 байт для поддержки 512-битных кривых.

### [4] Обновление сборочной конфигурации и TPM 1.2

- Добавлена линковка с `libgost` в `libtpms.pc.in`.
- Добавлен `configure.ac`: опция `use_openssl_functions_gost3410`.
- Обновлён `Makefile.am`: добавлены новые исходные файлы (`TpmEcc_Signature_GOST3410.c`, `TpmToGostSupport.c`) и заголовки.
- TPM 1.2 (tpm12): добавлены идентификаторы `TPM_ALG_MAGMA256` (`0x0B`) и `TPM_ALG_GRASSHOPPER256` (`0x0C`) в `tpm_constants.h`, обновлены `tpm_key.c`, `tpm_process.c`, `tpm_transport.c` для распознавания новых алгоритмов.
- Обновлена таблица compile-constants в `NVMarshal.c` (размер массива 122→133) для поддержки новых алгоритмов и кривых при миграции состояния.

---

## GOST-GO-TPM

**Репозиторий:** [github.com/IWillBurn/gost-go-tpm](https://github.com/IWillBurn/gost-go-tpm)


### [1] Добавлены константы ГОСТ-алгоритмов

**Файлы:** `tpm2/constants.go`, `legacy/tpm2/constants.go`, `tpm/constants.go`

**TPM 2.0 API (`tpm2/constants.go`):**

- `TPMAlgGOST3411256` = `0x0100`, `TPMAlgGOST3411512` = `0x0101` (хеширование)
- `TPMAlgMagma` = `0x0102`, `TPMAlgGrasshopper` = `0x0103` (симметричное шифрование)
- `TPMAlgGOST3410256` = `0x0104`, `TPMAlgGOST3410512` = `0x0105` (ЭЦП)
- Кривые ТК26: `TPMECCCurveTC26Gost3410256ParamSetA/B/C/D` (`0x0006`–`0x0009`), `TPMECCCurveTC26Gost3410512ParamSetA/B/C` (`0x000A`–`0x000C`)

**Legacy API (`legacy/tpm2/constants.go`):**

- Аналогичные константы: `AlgMagma`, `AlgGrasshopper`, `AlgGOST3410_256`, `AlgGOST3410_512`
- Кривые: `CurveTC26Gost3410256ParamSetA/B/C/D`, `CurveTC26Gost3410512ParamSetA/B/C`
- Исправлены значения `CurveBNP256` (`0x0010`) и `CurveBNP638` (`0x0011`) — ранее использовались iota, что конфликтовало с новыми кривыми.
- Добавлены строковые представления в `Algorithm.String()`.

**TPM 1.2 API (`tpm/constants.go`):**

- `AlgMagma256`, `AlgGrasshopper256` с маппингом в `AlgMap`.


### [2] Добавлена поддержка ГОСТ-типов в маршалинг/демаршалинг

**Файл:** `tpm2/structures.go`

**Симметричные алгоритмы:**

- `TPMUSymKeyBits`: добавлена обработка `TPMAlgMagma` и `TPMAlgGrasshopper` в `create()`, `get()`, `Sym()`. Добавлены методы `Magma()` и `Grasshopper()`.
- `TPMUSymMode`: аналогичные изменения — `create()`, `get()`, `Sym()`, методы `Magma()`, `Grasshopper()`.
- `TPMUSymDetails`: добавлена обработка `TPMAlgMagma` / `TPMAlgGrasshopper` как `TPMSEmpty`.

**Схемы подписи:**

- Новый тип `TPMSSigSchemeGOST3410` (alias `TPMSSchemeHash`).
- `TPMUSigScheme`: добавлена обработка `TPMAlgGOST3410256` / `TPMAlgGOST3410512` в `create()`, `get()`. Методы `GOST3410256()`, `GOST3410512()`.
- `TPMUAsymScheme`: добавлен `TPMSSigSchemeGOST3410` в type constraint `AsymSchemeContents`. Обработка в `create()`, `get()`. Методы `GOST3410256()`, `GOST3410512()`.
- `TPMUSignature`: добавлена обработка `TPMAlgGOST3410256` / `TPMAlgGOST3410512` как `TPMSSignatureECC`. Методы `GOST3410256()`, `GOST3410512()`.

**Legacy API (`legacy/tpm2/structures.go`):**

- `DecodeSignature`: добавлена обработка `AlgGOST3410_256` / `AlgGOST3410_512` — декодирование ECC-подписи (R, S).

---

## GOST-TPM
**Репозиторий:** [github.com/IWillBurn/gost-tpm](https://github.com/IWillBurn/gost-tpm)


### [1] Инициализация проекта и структура submodules

- Создан `.gitmodules` с подключением 4 submodules:
  - **gost-libtpms** — форк libtpms с ГОСТ-алгоритмами
  - **gost-swtpm** — форк swtpm
  - **gost-go-tpm** — форк go-tpm с ГОСТ-поддержкой
  - **gost-engine** — gost-engine (криптобиблиотека ГОСТ для OpenSSL)
- Создан `.gitignore` (IDE-файлы).
- Создан `README.md` с инструкциями по сборке и запуску.

### [2] Docker-инфраструктура

**Файлы:** `.docker/build.Dockerfile`, `.docker/tests.Dockerfile`, `.docker/scripts/`

**`build.Dockerfile` — многоэтапная сборка (multi-stage):**

- **Stage 1** (`build-swtpm`): сборка gost-engine → gost-libtpms → gost-swtpm на Ubuntu 22.04 с полным набором зависимостей.
- **Stage 2** (`runtime`): минимальный образ с установленными библиотеками, swtpm, tpm2-tools.
- Скрипт `run_build.sh`: инициализация swtpm (`swtpm_setup --profile name=custom:gost`), запуск swtpm socket, selftest.

**`tests.Dockerfile` — расширенная сборка для тестирования:**

- **Stage 1**: аналогична build.
- **Stage 2** (`build-tests`): сборка Go-тестов (`golang:1.25-alpine`, `go test -c`).
- **Stage 3** (`runtime`): запуск swtpm + выполнение Go-тестов.
- Скрипт `run_tests.sh`: инициализация swtpm, запуск тестового бинарника.

### [3] Интеграционные тесты (Go)

**Директория:** `go-tests/`

- **Модуль Go** (`go.mod`): зависимость от `github.com/IWillBurn/gost-go-tpm`, `testify`.
- **Makefile** (`scripts/Makefile`): компиляция тестового бинарника для `linux/amd64`.

**Утилиты (`tests/crypto_utils.go`):**

- `mustCreatePrimary` — создание RSA primary key (owner hierarchy).
- `createSymmetricKey`, `loadKey`, `encryptSymmetric`, `decryptSymmetric` — операции с симметричными ключами через `EncryptDecrypt2`.
- `hashWithTPM`, `hashSequenceWithChunks` — хеширование (одиночное и последовательное с context save/restore).
- `getCapabilityAlg` — проверка наличия алгоритма через `GetCapability`.
- Функции для ГОСТ-подписи: `createGOST3410SigningKey`, `loadGOST3410SigningKey`, `signGOST3410`, `verifySignatureByHandle`.

**Тесты:**

- **`Test_HasAlgs`** (`has_algs_test.go`): проверяет, что TPM сообщает о поддержке всех 6 ГОСТ-алгоритмов: `GOST3411-256`, `GOST3411-512`, `Magma`, `Grasshopper`, `GOST3410-256`, `GOST3410-512`.

- **`Test_Hash`** (`hash_test.go`): проверяет корректность хеширования ГОСТ Р 34.11-2012 по эталонным векторам из RFC 6986 (4 тест-кейса: M1/M2 для 256 и 512 бит).

- **`Test_NVMarshaling_Hashing`** (`marshaling_test.go`): проверяет корректность маршалинга хеш-контекста: хеширование данных по частям с `HashSequenceStart` → `SequenceUpdate` + `ContextSave` / `ContextLoad` → `SequenceComplete` должно давать тот же результат, что и хеширование целиком. Тестируется для `GOST3411-256` и `GOST3411-512`.

- **`Test_CryptDecrypt`** (`crypt_test.go`): проверяет симметричное шифрование/дешифрование через `EncryptDecrypt2`. 4 тест-кейса:
  - Magma-CBC, Magma-CTR
  - Grasshopper-CBC, Grasshopper-CTR

  Создаётся симметричный ключ, шифруются данные, дешифруются — проверяется совпадение с оригиналом.

- **`Test_ECDSASign_Verify`** (`sign_test.go`): проверяет подпись и верификацию ГОСТ Р 34.10-2012. 7 тест-кейсов по всем поддерживаемым кривым:
  - `gost3410-256` + paramSetA/B/C/D
  - `gost3410-512` + paramSetA/B/C

  Создаётся ECC-ключ с ГОСТ-схемой, подписывается дайджест, верифицируется подпись.