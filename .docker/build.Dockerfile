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
    make \
    gcc \
    g++ \
    libssl-dev \
    libjson-glib-dev \
    libgmp3-dev \
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
# STAGE 3: Run
# ==============================================================================
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y -f \
    libssl3 \
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

# Copy swtpm
COPY --from=build-swtpm /usr/bin/swtpm* /usr/bin/
COPY --from=build-swtpm /usr/share/swtpm /usr/share/swtpm
COPY --from=build-swtpm /usr/include/libtpms /usr/include/libtpms
COPY --from=build-swtpm /usr/include/swtpm /usr/include/swtpm
COPY --from=build-swtpm /usr/lib/swtpm /usr/lib/swtpm
COPY --from=build-swtpm /usr/lib/libtpms* /usr/lib/

# Copy run script
COPY ./.docker/scripts/run_build.sh /scripts/
RUN chmod +X /scripts/run_build.sh

# Run
CMD /scripts/run_build.sh