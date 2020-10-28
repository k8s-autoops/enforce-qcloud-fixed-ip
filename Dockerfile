FROM golang:1.14 AS builder
ENV GOPROXY https://goproxy.io
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /admission-httpscat

FROM alpine:3.12
COPY --from=builder /admission-httpscat /admission-httpscat
CMD ["/admission-httpscat"]