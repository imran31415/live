FROM golang:1.14-alpine AS builder
RUN apk add --update --no-cache ca-certificates git
WORKDIR /project
COPY go.mod .
COPY go.sum .
RUN go mod download
# COPY the source code as the last step
COPY . .
RUN export GOPROXY=https://goproxy.cn && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main main.go


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /project/main /main
EXPOSE 9033

ENTRYPOINT ["/main"]





