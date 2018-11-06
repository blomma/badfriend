FROM arm32v6/alpine

COPY bin/linux-arm-7-badfriend /linux-arm-7-badfriend

EXPOSE 8000

CMD ["/linux-arm-7-badfriend"]
