FROM golang:1-bullseye as builder0

WORKDIR /app-build
COPY . /app-build
RUN go build github.com/mailpond/mailpond-2/cmd/mailpond-serviced   \
	&& mkdir /app-build/binaries                                    \
	&& mv mailpond-serviced /app-build/binaries


FROM debian:11

RUN apt-get update                      \
	&& apt-get -y dist-upgrade          \
	&& apt-get -y clean                 \
	&& mkdir -p /var/lib/mailpond

COPY --from=builder0 /app-build/binaries/*  /opt/mailpond/bin/

CMD ["/opt/mailpond/bin/mailpond-serviced", "-storagePath", "/var/lib/mailpond"]
