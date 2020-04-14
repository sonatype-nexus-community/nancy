FROM alpine:latest as builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/cache/apk/* 

COPY nancy /

#--------------------------------
# Deployment Image
#--------------------------------
FROM scratch

#Import from builder image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /nancy /nancy

ENTRYPOINT [ "/nancy" ]
