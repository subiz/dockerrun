FROM alpine:3.8
COPY dockerun /usr/local/bin/
ENTRYPOINT dockerun
