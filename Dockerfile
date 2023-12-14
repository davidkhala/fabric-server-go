FROM golang:1.16 as golang
LABEL org.opencontainers.image.source = "https://github.com/davidkhala/fabric-server-go"
WORKDIR /root
COPY .. .
RUN go mod vendor
RUN go build -o /root/app ./main
CMD ["/root/app"]
# TODO why alpine not working


