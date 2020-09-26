ARG BUILD_DIR=/go/src/app
ARG SERVICE_NAME=di-velocity

FROM golang:1.14 AS build
ARG BUILD_DIR
ARG SERVICE_NAME

WORKDIR ${BUILD_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ${SERVICE_NAME} -v ./pkg/nats

FROM ubuntu AS run
ARG BUILD_DIR
ARG SERVICE_NAME

RUN useradd -ms /bin/bash docker
USER docker

WORKDIR /app
ENV PATH="/app:${PATH}"

COPY --from=build --chown=docker:docker ${BUILD_DIR} .

CMD ["./di-velocity"]
