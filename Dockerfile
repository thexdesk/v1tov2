FROM golang:1.9.3-alpine3.7
RUN apk add -U gcc
COPY . /go/src/github.com/hinshun/v1tov2/
RUN CGO_ENABLED=0 GOOS=linux go build github.com/hinshun/v1tov2/cmd/v1tov2

FROM alpine:3.21.2
COPY --from=0 /go/v1tov2 /bin
ENTRYPOINT ["v1tov2"]
