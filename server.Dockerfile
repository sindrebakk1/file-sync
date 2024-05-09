FROM golang:1.22-alpine3.19 AS base
LABEL authors="Sindre"

WORKDIR /src

COPY go.work go.work.sum ./
COPY ./cmd/server/go.* ./cmd/server/
COPY ./cmd/client/go.* ./cmd/client/
COPY ./pkg/constants/go.* ./pkg/constants/
COPY ./pkg/enums/go.* ./pkg/enums/
COPY ./pkg/models/go.* ./pkg/models/
COPY ./pkg/integration/go.* ./pkg/integration/

RUN go mod download -x

FROM base AS build

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server -v server

FROM scratch AS server

COPY cmd/server/_certs /root/certs
COPY config.json config.dev.json /etc/filesync/
COPY --from=build /bin/server /bin/server

EXPOSE 443

ENTRYPOINT ["/bin/server"]