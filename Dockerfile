FROM golang:1.17-alpine as build-container
COPY . /root/
WORKDIR /root
RUN CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a -ldflags '-s -w' && \
go get github.com/google/go-licenses && \
go-licenses save ./... --save_path=legal

FROM gcr.io/distroless/static
COPY --from=build-container --chown=nobody:nobody /root/webserver /app/webserver
COPY --from=build-container --chown=nobody:nobody /root/legal /app/legal
USER nobody
EXPOSE 8080 8081
ENTRYPOINT ["/app/webserver"]
CMD []
