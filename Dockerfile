FROM golang:1.23

WORKDIR ${GOPATH}/medods-task/
COPY . ${GOPATH}/medods-task/

RUN go build -o /build ./cmd

EXPOSE 8080

CMD ["/build"]