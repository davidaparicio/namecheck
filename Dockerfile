FROM alpine:3.22 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
#ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY namecheck /usr/local/bin/namecheck
USER 65534:65534
ENTRYPOINT [ "/usr/local/bin/namecheck" ]
