FROM ubuntu:18.04 AS build

ENV LIBVIPS_VERSION="8.7.0" \
    LIBVIPS_DOWNLOAD_URL="https://github.com/libvips/libvips/releases/download/v${LIBVIPS_VERSION}/vips-${LIBVIPS_VERSION}.tar.gz" \
    LIBVIPS_DOWNLOAD_SHA256="1dfe664fa3d8ad714bbd15a36627992effd150ddabd7523931f077b3926d736d" \
    GOLANG_VERSION="1.11.4" \
    GOLANG_DOWNLOAD_URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    GOLANG_DOWNLOAD_SHA256="bc20d46d5f5cf274b76fbd29e1912cd9e4feb0893cc3bcdd9e528e6ab966525c" \
    DEBIAN_FRONTEND="noninteractive" \
    PACKAGE="github.com/ameliaikeda/sakura"

# Install required libraries
RUN \
    apt-get update && \
    apt-get install -y \
    ca-certificates \
    automake build-essential curl gcc git libc6-dev make \
    gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg-turbo8-dev libpng12-dev \
    libwebp-dev libtiff5-dev libgif-dev libexif-dev libxml2-dev libpoppler-glib-dev \
    swig libmagickwand-dev libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev \
    libgsf-1-dev fftw3-dev liborc-0.4-dev librsvg2-dev && \

    # update ca-certificates
    update-ca-certificates && \

    cd /tmp && \

    # verify the download before unpacking
    curl -fsSL "$LIBVIPS_DOWNLOAD_URL" -o libvips.tar.gz && \
    echo "$LIBVIPS_DOWNLOAD_SHA256 libvips.tar.gz" | sha256sum -c - && \

    # unpack and build
    tar zvxf libvips.tar.gz && \
    cd /tmp/libvips && \
    ./configure --enable-debug=no --without-python $1 && \
    make && \
    make install && \
    ldconfig && \

    # download and verify golang
    curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz  && \
    echo "$GOLANG_DOWNLOAD_SHA256 golang.tar.gz" | sha256sum -c - && \
    tar -C /usr/local -xzf golang.tar.gz && \
    rm golang.tar.gz && \

    # clean up.
    apt-get remove -y curl automake build-essential libc6-dev gcc && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


# copy our repo code over.
WORKDIR /code
COPY . .

RUN go mod vendor -v && go build -o /bin/sakura cmd/sakura/main.go

FROM ubuntu:18.04 as production

LABEL maintainer="Amelia Ikeda <amelia@lolibrary.org>" \
    licence="BSD-3-Clause" \
    issues="https://github.com/ameliaikeda/sakura/issues" \
    homepage="https://github.com/ameliaikeda/sakura"

# all config environment variables.
ENV DEBIAN_FRONTEND="noninteractive" \
    HOST="" \
    PORT="" \
    MAX_HEIGHT="" \
    IMAGE_BUCKET="" \
    THUMBNAIL_BUCKET="" \
    AWS_ENDPOINT="" \
    AWS_ACCESS_KEY_ID="" \
    AWS_SECRET_KEY=""


RUN \
    # Install runtime dependencies
    apt-get update -y && \
    apt-get install --no-install-recommends -y \
        libglib2.0-0 libjpeg-turbo8 libpng12-0 libopenexr22 \
        libwebp5 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
        libmagickwand-6.q16-2 libpango1.0-0 libmatio2 libopenslide0 \
        libgsf-1-114 fftw3 liborc-0.4 librsvg2-2 libcfitsio2 && \
    # Clean up
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /usr/local/lib /usr/local/lib
RUN ldconfig
COPY --from=build /bin/sakura /usr/local/bin/sakura
COPY --from=build /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/usr/local/bin/sakura"]
