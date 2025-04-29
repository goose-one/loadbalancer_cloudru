FROM golang:1.24 AS build-stage

WORKDIR /app

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main/


FROM alpine:3 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app /

ENV CONFIG_FILE=./configs/prod.yaml

EXPOSE 8080
#USER nonroot:nonroot
ENTRYPOINT ["/main"]
