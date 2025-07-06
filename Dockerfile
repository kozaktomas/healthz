FROM golang:1.24 AS backend
ENV CGO_ENABLED=0
ADD . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -v -o healthz .


FROM alpine:3
RUN apk update && \
    apk add openssl tzdata && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile

COPY --from=backend /app/healthz /app/healthz

RUN chown nobody /app/healthz \
    && chmod 500 /app/healthz

USER nobody

ENTRYPOINT ["/app/healthz"]
