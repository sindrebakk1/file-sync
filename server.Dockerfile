FROM golang:1.22-alpine3.19 AS base
LABEL authors="Sindre"

WORKDIR /src

COPY . .
RUN go work sync

FROM base AS build

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server server

FROM scratch AS server

COPY cmd/server/_certs /root/certs
COPY config.json config.dev.json /etc/filesync/
COPY --from=build /bin/server /bin/server

EXPOSE 443

ENTRYPOINT ["/bin/server"]