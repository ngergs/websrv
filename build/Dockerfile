FROM gcr.io/distroless/static-debian13:nonroot
COPY linux/amd64/websrv /app/websrv
COPY legal /app/legal
USER nobody
EXPOSE 8080 8081
ENTRYPOINT ["/app/websrv"]
CMD []
