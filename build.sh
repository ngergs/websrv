#!/bin/bash
docker pull golang:1.17-alpine &&\
docker pull gcr.io/distroless/static:latest &&\
docker build --tag=selfenergy/webserver . &&\
docker push selfenergy/webserver
