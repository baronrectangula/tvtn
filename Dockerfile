# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build

WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local

COPY . .
ARG CMD=rest-server
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/app "./cmd/${CMD}"

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/app /app/server
WORKDIR /app
USER 65532:65532
ENTRYPOINT ["/app/server"]
