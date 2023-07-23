FROM golang:1.21-rc-alpine AS builder

ARG UID=1000
ARG GID=1000

RUN apk add upx shadow

RUN groupadd --gid ${GID} bobthebuilder && useradd -m --gid ${GID} --uid ${UID} bobthebuilder

RUN mkdir /build
RUN chown -Rv bobthebuilder:bobthebuilder /build

COPY . /workspace
RUN chown -Rv bobthebuilder:bobthebuilder /workspace

USER bobthebuilder
WORKDIR /workspace

RUN [ "go", "build", "-ldflags", "-s -w", "-o", "/build/docker_launcher", "." ]

RUN [ "upx", "/build/docker_launcher" ]

FROM alpine:latest

COPY --from=builder /build/docker_launcher /usr/local/bin/docker_launcher

ENTRYPOINT [ "/usr/local/bin/docker_launcher" ]
