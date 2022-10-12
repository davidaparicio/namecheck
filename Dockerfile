FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
#ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY namecheck /usr/local/bin/namecheck
ENTRYPOINT [ "/usr/local/bin/namecheck" ]