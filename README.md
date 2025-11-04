# gost-tpm

## PREPARING

### Load submodules
```shell
git submodule update --init --recursive
```

## BUILD CASE

### Build
```shell
docker build -f ./.docker/build.Dockerfile -t gost-tpm .
```

### Run
```shell
docker run gost-tpm
```

## TEST CASE

### Build
```shell
docker build -f ./.docker/tests.Dockerfile -t gost-tpm-tests .
```

### Run
```shell
docker run gost-tpm-tests
```