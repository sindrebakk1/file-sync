FROM golang:1.22 as sync
LABEL authors="Sindre"

WORKDIR go/src/app
COPY . .

RUN go work sync

FROM sync as build

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/server server

#FROM build as test
#
#RUN  go test server/...

FROM gcr.io/distroless/base-debian11 as build-release

COPY --from=sync /app/cmd/server/certs /root/certs

EXPOSE 443

ENTRYPOINT ["/server"]