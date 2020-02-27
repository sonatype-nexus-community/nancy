FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    rm -rf /var/cache/apk/* 

COPY nancy /

ENTRYPOINT [ "/nancy" ]
