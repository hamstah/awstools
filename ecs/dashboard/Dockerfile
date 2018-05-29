FROM alpine:3.7

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /usr/share

EXPOSE 8000
COPY ecs-dashboard /bin/ecs-dashboard

CMD ["/bin/ecs-dashboard"]