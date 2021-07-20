FROM golang:1.16-buster AS build

RUN apt update \
 && apt install -y make git \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ./go.* ./

RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY Makefile ./

ARG version

RUN BIN="/csi" VERSION="$version" make controller
RUN BIN="/csi" VERSION="$version" make node

###########################################

FROM ubuntu:18.04

LABEL org.opencontainers.image.title="Seagate Exos X CSI" \
      org.opencontainers.image.description="A dynamic persistent volume provisioner for Seagate Exos X based storage systems." \
      org.opencontainers.image.url="https://github.com/Seagate/seagate-exos-x-csi" \
      org.opencontainers.image.source="https://github.com/Seagate/seagate-exos-x-csi/blob/master/Dockerfile" \
      org.opencontainers.image.documentation="https://github.com/Seagate/seagate-exos-x-csi/blob/master/README.md" \
      org.opencontainers.image.licenses="Apache 2.0"

COPY --from=build /seagate-exos-x-csi-* /usr/local/bin/

ENV PATH="${PATH}:/lib/udev"

CMD [ "/usr/local/bin/seagate-exos-x-csi-controller" ]

ARG version
ARG vcs_ref
ARG build_date
LABEL org.opencontainers.image.version="$version" \
      org.opencontainers.image.revision="$vcs_ref" \
      org.opencontainers.image.created="$build_date"
