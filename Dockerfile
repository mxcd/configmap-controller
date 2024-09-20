FROM golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /usr/src

COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum

RUN go mod download

# Copy the go source
COPY cmd/controller/main.go /usr/src/cmd/controller/main.go
COPY internal /usr/src/internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o controller /usr/src/cmd/controller/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /usr/src/controller .
USER 65532:65532

ENTRYPOINT ["/controller"]
