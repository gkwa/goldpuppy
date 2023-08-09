FROM ubuntu:latest

RUN apt-get update
RUN apt-get -y install curl

RUN curl -sSL https://github.com/taylormonacelli/goldpuppy/releases/latest/download/goldpuppy_Linux_x86_64.tar.gz | tar -C /usr/local/bin --no-same-owner -xz goldpuppy
RUN goldpuppy --help

