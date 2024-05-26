FROM ubuntu:latest
LABEL authors="khristina"

ENTRYPOINT ["top", "-b"]