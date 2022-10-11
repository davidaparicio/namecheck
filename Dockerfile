FROM scratch
COPY ca-certificates.crt /etc/ssl/certs/
COPY namecheck /usr/local/bin/namecheck
ENTRYPOINT [ "/usr/local/bin/namecheck" ]