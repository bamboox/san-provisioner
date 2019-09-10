#FROM golang:1.12.4 as builder
FROM registry.cn-hangzhou.aliyuncs.com/bamboo/golang:1.12.4 as builder

ARG DIR=$GOPATH/src/github.com/kubernetes-sigs/san-client

WORKDIR $DIR

RUN mkdir -p $DIR

COPY / $DIR

ENV  GO111MODULE=on
# setting go mod proxy, proxy run on onebox
ENV  GOPROXY=https://goproxy.io
#RUN  go mod tidy
#RUN  go mod vendor
#ENV  GO111MODULE=off

RUN  go build -o main -ldflags '-s -w' -v $DIR/cmd/san-client-provisioner

FROM alpine:3.9

ARG APK_MIRROR=mirrors.aliyun.com
RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories

RUN apk add --no-cache libc6-compat
## DON'T modify above, as it's common for all Alpine based parent and docker caching layer will used

CMD ["./main"]

COPY --from=builder /go/src/github.com/kubernetes-sigs/san-client/main .


