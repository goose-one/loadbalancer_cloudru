FROM golang:1.24 AS build-stage

WORKDIR /app

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build .


FROM alpine:3 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app /

EXPOSE 8080
#USER nonroot:nonroot
ENTRYPOINT ["/server"]
