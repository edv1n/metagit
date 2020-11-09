FROM golang:1.15.3-alpine as build
WORKDIR /metagit
COPY . .
RUN go build -ldflags="-extldflags=-static"

FROM alpine:latest
COPY --from=build /metagit/metagit /usr/local/bin/
CMD ["metagit", "serve"]