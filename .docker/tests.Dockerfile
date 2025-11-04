# ==============================================================================
# STAGE 1: Build SWTPM
# ==============================================================================
FROM ubuntu:22.04 AS build-swtpm

# Download dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    autoconf \
    automake \
    libtool \
    cmake \
    gcc \
    g++ \
    libssl-dev \
    libjson-glib-dev \
    libgmp3-dev \
    build-essential \
    expect \
    socat \
    python3 \
    python3-setuptools \
    python3-cryptography \
    python3-twisted \
    pkg-config \
    libtasn1-6-dev \
    net-tools \
    iproute2 \
    gawk \
    libseccomp-dev \
    dos2unix \
    libfuse-dev \
    libglib2.0-dev \
    libgnutls28-dev \
    trousers \
    libtspi-dev \
    make \
    tpm2-tools \
    && rm -rf /var/lib/apt/lists/*

# Copy sources
COPY ./gost-libtpms /src/gost-libtpms
COPY ./gost-swtpm /src/gost-swtpm
COPY ./gost-engine /src/gost-engine

# Build gost-engine
WORKDIR /src/gost-engine

RUN mkdir -p /usr/include/gost-engine/
RUN cp *.h /usr/include/gost-engine/
RUN mkdir build

WORKDIR /src/gost-engine/build
RUN cmake -DCMAKE_BUILD_TYPE=Release ..
RUN cmake --build . --target install --config Release
RUN cp ./bin/gost.so /usr/lib/
RUN cp ./bin/libgost.so /usr/lib/

RUN ldconfig

# Build gost-libtpms
WORKDIR /src/gost-libtpms
RUN find . -type f -exec dos2unix --quiet --safe {} \+ || true
RUN chmod +x *.sh
RUN ./autogen.sh --prefix=/usr --with-openssl --with-tpm2
RUN make -j$(nproc)
RUN make install

# Build gost-swtpm
WORKDIR /src/gost-swtpm
RUN ls -la  # Отладка
RUN if [ -f autogen.sh ]; then cat autogen.sh; else echo "autogen.sh not found"; fi
RUN if [ -f bootstrap.sh ]; then cat bootstrap.sh; else echo "bootstrap.sh not found"; fi
RUN find . -type f -exec dos2unix --quiet --safe {} \+ || true
RUN chmod +x *.sh
RUN ./autogen.sh --prefix=/usr --with-openssl
RUN make -j$(nproc)
RUN make install

# ==============================================================================
# STAGE 2: Build tests
# ==============================================================================
FROM golang:1.25-alpine AS build-tests

# Download dependencies
RUN apk add --no-cache make

WORKDIR /src/go-tests

# Copy sources
COPY ./go-tests .
RUN go mod download
RUN make -f ./scripts/Makefile build-tests

# ==============================================================================
# STAGE 3: Run
# ==============================================================================
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y -f \
    libssl3 \
    cmake \
    libjson-glib-1.0-0 \
    libgmp10 \
    socat \
    libtasn1-6 \
    libseccomp2 \
    libglib2.0-0 \
    libgnutls30 \
    libfuse3-3 \
    trousers \
    tpm2-tools \
    tpm2-abrmd \
    && rm -rf /var/lib/apt/lists/*

# Copy gost-engine
COPY --from=build-swtpm /src/gost-engine/build/bin/gost.so /usr/lib/x86_64-linux-gnu/engines-3/

COPY --from=build-swtpm /src/gost-engine/build/bin/gost.so /usr/lib/
COPY --from=build-swtpm /src/gost-engine/build/bin/libgost.so /usr/lib/

# Copy swtpm
COPY --from=build-swtpm /usr/bin/swtpm* /usr/bin/
COPY --from=build-swtpm /usr/share/swtpm /usr/share/swtpm
COPY --from=build-swtpm /usr/include/libtpms /usr/include/libtpms
COPY --from=build-swtpm /usr/include/swtpm /usr/include/swtpm
COPY --from=build-swtpm /usr/lib/swtpm /usr/lib/swtpm
COPY --from=build-swtpm /usr/lib/libtpms* /usr/lib/

# Copy tests
COPY --from=build-tests /src/go-tests/build/go-tests /tests/go-tests

# Copy run script
COPY ./.docker/scripts/run_tests.sh /scripts/
RUN chmod +x /scripts/run_tests.sh

# Run
CMD /scripts/run_tests.sh