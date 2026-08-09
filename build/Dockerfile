FROM gcr.io/distroless/static-debian12:nonroot
COPY websrv /app/websrv
COPY legal /app/legal
USER nobody
EXPOSE 8080 8081
ENTRYPOINT ["/app/websrv"]
CMD []
