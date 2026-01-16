FROM alpine:latest@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62 as certs
RUN apk --update add ca-certificates

FROM scratch
#ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY namecheck /usr/local/bin/namecheck
ENTRYPOINT [ "/usr/local/bin/namecheck" ]