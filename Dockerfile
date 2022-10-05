############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git gcc libc-dev openssl && go install github.com/gobuffalo/packr/packr@latest
WORKDIR $GOPATH/src/securitybunker/databunker/src/
COPY src/go.mod ./deps
RUN cat ./deps | grep -v storage > ./go.mod && go mod download
COPY . $GOPATH/src/securitybunker/databunker/
#RUN echo "tidy " && go get -u && go mod tidy && cat ./go.mod
# Fetch dependencies.
# Using go get.
RUN go get -d -v && \
    packr && \
    go build -o /go/bin/databunker && \
    packr clean
############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /bin/sh /bin/busybox /usr/bin/openssl /bin/
COPY --from=builder /lib/ld* /lib/libssl.* /lib/libcrypto.* /lib/
COPY --from=builder /etc/group /etc/
COPY --from=builder /etc/ssl /etc/ssl

COPY databunker.yaml /databunker/conf/
RUN /bin/busybox mkdir -p /databunker/data && \
    /bin/busybox mkdir -p /databunker/certs && \
    /bin/busybox ln -s /bin/busybox /bin/addgroup && \
    /bin/busybox ln -s /bin/busybox /bin/adduser && \
    /bin/busybox ln -s /bin/busybox /bin/chown && \
    /bin/busybox touch /etc/passwd && \
    /bin/busybox mkdir -p /tmp && \
    /bin/busybox chmod 0777 /tmp && \
    addgroup -S appgroup && adduser --no-create-home -S appuser -G appgroup && \
    chown appuser:appgroup /databunker/data
USER appuser
COPY --from=builder /go/bin/databunker /go/src/securitybunker/databunker/run.sh /go/src/securitybunker/databunker/health-check.sh /databunker/bin/
EXPOSE 3000
HEALTHCHECK --interval=5s --timeout=3s --start-period=33s --retries=3 CMD /databunker/bin/health-check.sh
ENTRYPOINT ["/bin/sh", "/databunker/bin/run.sh"]
#CMD ["/bin/sh", "-x", "-c", "/go/bin/databunker -init"]
