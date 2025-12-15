FROM ubuntu:latest
LABEL authors="denisdavydov"

ENTRYPOINT ["top", "-b"]