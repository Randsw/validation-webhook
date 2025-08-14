# Build the kubeinfo binary
FROM golang:1.25 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o validatewebhook main.go

# Use distroless as minimal base image to package the kubeinfo binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/validatewebhook .
USER 65532:65532

CMD ["./validatewebhook", "--tls-cert", "/etc/opt/tls.crt", "--tls-key", "/etc/opt/tls.key", "--port", "3000"]