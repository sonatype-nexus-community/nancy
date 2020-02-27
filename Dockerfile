FROM alpine:latest

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/cache/apk/* 

COPY nancy /

ENTRYPOINT [ "/nancy" ]
