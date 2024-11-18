FROM golang:latest AS image_golang
RUN mkdir -p /go/src/pipelinesf
WORKDIR /go/src/pipelinesf
ADD main/main.go .
ADD go.mod .
RUN go env -w GO111MODULE=auto
RUN go install .
 
FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Pipeline panarinsf<panarinsf@sf.ru>"
WORKDIR /root/
COPY --from=image_golang /go/bin/pipelinesf .
ENTRYPOINT ["./pipelinesf"]