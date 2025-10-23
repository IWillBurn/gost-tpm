# gost-tpm

## BUILD CASE

### Build
```shell
docker build -f ./.docker/build.Dockerfile -t gost-tpm .
```

### Run
```shell
docker run gost-tpm
```

## TEST STUBS CASE

### Build
```shell
docker build -f ./.docker/tests.Dockerfile -t gost-tpm-stub-tests .
```

### Run
```shell
docker run gost-tpm-stub-tests
```