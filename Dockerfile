FROM golang:1.17-alpine as build-container
COPY . /root/
WORKDIR /root
RUN CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a --ldflags '-s -w'

FROM gcr.io/distroless/static
COPY --from=build-container --chown=nobody:nobody /root/webserver /app/webserver
USER nobody
EXPOSE 8080 8081
ENTRYPOINT ["/app/webserver"]
CMD []
