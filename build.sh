#!/usr/bin/env bash

# FOR DEV
# Build Namecheck CLI
#goreleaser build --single-target --snapshot -f .goreleaser.yaml --rm-dist
# Build Namecheck SERVER
goreleaser build --single-target --snapshot -f .goreleaser_server.yaml --rm-dist

# RUN Swagger UI
#docker run -p 47101:8080 -e SWAGGER_JSON_URL=https://raw.githubusercontent.com/davidaparicio/namecheck/main/api/swagger.yaml swaggerapi/swagger-ui
#docker run -p 1337:8080 -e BASE_URL=/swagger -e SWAGGER_JSON=/foo/swagger.yaml -v /api:/foo swaggerapi/swagger-ui