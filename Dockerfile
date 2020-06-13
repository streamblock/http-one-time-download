FROM golang:1.14-alpine

RUN set -x ; \
  addgroup -g 82 -S www-data ; \
  adduser -u 82 -D -S -G www-data www-data
RUN apk add git gcc libc-dev mailcap

# /var/www/share refers to server/upload_path in config.yml
RUN mkdir -p /var/www/share && chown www-data:www-data /var/www/share && chmod 777 /var/www/share

WORKDIR /app
COPY server/ ./
RUN go get gopkg.in/yaml.v2
RUN go build -o server .

VOLUME /var/www/share
EXPOSE 8080/tcp
CMD ["./server"]
