FROM golang:latest

ARG UID=1000
ARG GID=1000

RUN addgroup --gid ${GID} bobthebuilder && adduser --gid ${GID} --uid ${UID} bobthebuilder

RUN mkdir /build
RUN chown -Rv bobthebuilder:bobthebuilder /build

COPY . /workspace
RUN chown -Rv bobthebuilder:bobthebuilder /workspace

USER bobthebuilder
WORKDIR /workspace

CMD [ "go", "build", "-o", "/build/docker_launcher", "." ]
