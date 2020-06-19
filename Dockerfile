FROM golang:1.13.12
WORKDIR /go/src/github.com/tynany/junos_exporter
COPY . /go/src/github.com/tynany/junos_exporter
RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine:3.12.0
WORKDIR /app
COPY --from=0 /go/src/github.com/tynany/junos_exporter/junos_exporter .
EXPOSE 9347
CMD ["./junos_exporter", "--config.path=/config.yaml"]
