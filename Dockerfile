FROM alpine:latest
MAINTAINER Adam Carruthers <adam@bitjutsu.ca>

RUN apk update
RUN apk add iptables

COPY target/vr /usr/bin/vr

EXPOSE 8080

ENTRYPOINT vr
