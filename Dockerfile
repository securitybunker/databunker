##############################
# STEP 1 build executable file
##############################
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git gcc libc-dev openssl && go install github.com/gobuffalo/packr/packr@latest
WORKDIR /go/src/securitybunker/databunker/src/
COPY src/go.mod ./deps
RUN cat ./deps | grep -v storage | grep -v utils | grep -v audit > ./go.mod && go mod download
COPY . /go/src/securitybunker/databunker/
WORKDIR /go/src/securitybunker/databunker/
#RUN echo "tidy " && go get -u && go mod tidy && cat ./go.mod
RUN ./build.sh
#################################
# STEP 2 create a small container
#################################
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
VOLUME ["/databunker/data"]
USER appuser
COPY --from=builder /go/src/securitybunker/databunker/databunker /go/src/securitybunker/databunker/run.sh /go/src/securitybunker/databunker/health-check.sh /databunker/bin/
EXPOSE 3000
HEALTHCHECK --interval=5s --timeout=3s --start-period=33s --retries=3 CMD /databunker/bin/health-check.sh
ENTRYPOINT ["/bin/sh", "/databunker/bin/run.sh"]
#CMD ["/bin/sh", "-x", "-c", "/go/bin/databunker -init"]
