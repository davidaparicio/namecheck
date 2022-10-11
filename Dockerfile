FROM scratch
COPY namecheck /usr/local/bin/namecheck
ENTRYPOINT [ "/usr/local/bin/namecheck" ]