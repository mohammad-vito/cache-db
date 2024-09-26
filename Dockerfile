FROM golang:1.22.6 as go
# Copy the source code into the container.
COPY . /service

# Build the rest-api binary.
WORKDIR /service/cmd/rest

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn

RUN go mod tidy
RUN go mod vendor

RUN go build main.go


FROM alpine:3.9
RUN apk add libc6-compat

COPY --from=go /service/cmd/rest /service/rest-api

WORKDIR /service
