FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -o /out/app ./cmd/main.go

FROM scratch
USER 10001:10001
COPY --from=builder /out/app /app
ENTRYPOINT ["/app"]
