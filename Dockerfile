FROM golang:1.22
WORKDIR /go/src/github.com/tynany/junos_exporter
COPY . /go/src/github.com/tynany/junos_exporter
RUN make setup_promu
RUN ./promu build
RUN ls -lah

FROM alpine:3.20.3
WORKDIR /app
COPY --from=0 /go/src/github.com/tynany/junos_exporter/junos_exporter .
EXPOSE 9347
CMD ["./junos_exporter", "--config.path=/config.yaml"]
