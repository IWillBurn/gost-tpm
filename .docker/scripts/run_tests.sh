#!/bin/bash

mkdir -p /var/run/swtpm
mkdir -p /tpm/state
mkdir -p /tpm/socket

echo "[INFO]: Starting setup swtpm..."
swtpm_setup \
    --tpm-state "/tpm/state/" \
    --tpm2 \
    --profile name=custom:gost \
    --pcr-banks sha256 \
    --create-ek-cert \
    --create-platform-cert \
    --lock-nvram \
    --overwrite > /var/log/swtpm_setup.log 2>&1 &
sleep 5
echo "[INFO]: Setup swtpm is complete!"

echo "[INFO]: Starting swtpm..."
swtpm socket -d \
    --tpm2 \
    --tpmstate dir=/tpm/state/ \
    --profile name=custom:gost \
    --ctrl type=unixio,path=/tpm/socket/swtpm-socket.ctrl \
    --server type=unixio,path=/tpm/socket/swtpm-socket \
    --flags startup-clear \
    --log level=20 > /var/log/swtpm_socket.log 2>&1 &
sleep 5
echo "[INFO]: Swtpm started!"

echo "[INFO]: Starting swtpm selftest..."
tpm2_selftest --tcti="swtpm:path=/tpm/socket/swtpm-socket"
echo "[INFO]: Selftest swtpm is complete!"

echo "[INFO]: Starting tests..."
/tests/go-tests -test.v
echo "[INFO]: Tests complete!"