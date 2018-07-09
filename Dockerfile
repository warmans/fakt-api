FROM alpine:latest

#enable CGO (required for sqlite) and tzdata (required for time.Location)
RUN apk update && \
	apk --no-cache add ca-certificates &&\
	update-ca-certificates && \
	apk add openssl && \
	apk add --update curl gnupg tzdata && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub &&\
    wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.23-r3/glibc-2.23-r3.apk && apk add glibc-2.23-r3.apk

RUN mkdir -p /opt/fakt-api/static && \
    mkdir -p /opt/fakt-api/db && \
    mkdir -p /var/log/fakt-api && \
    mkdir -p /opt/fakt-api/migrations

COPY ./.build /opt/fakt-api
COPY ./docker-entrypoint.sh /opt/fakt-api
COPY ./migrations/* /opt/fakt-api/migrations/.

EXPOSE 8080
ENTRYPOINT ["/opt/fakt-api/docker-entrypoint.sh"]
CMD ["fakt-api"]