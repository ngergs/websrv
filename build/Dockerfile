FROM gcr.io/distroless/static
COPY websrv-linux-amd64 /app/websrv
COPY legal /app/legal
USER nobody
EXPOSE 8080 8081
ENTRYPOINT ["/app/websrv"]
CMD []
