FROM golang:alpine AS build_base
#ENV GOARCH arm64
#ENV GOARCH amd64
RUN apk add --no-cache git gcc ca-certificates libc-dev \
&& mkdir -p /go/src/github.com/xiaoxinpro/ \
&& cd /go/src/github.com/xiaoxinpro/ \
&& git clone https://github.com/xiaoxinpro/speedtest-go-zh.git
WORKDIR /go/src/github.com/xiaoxinpro/speedtest-go-zh
RUN go get ./ && go build -ldflags "-w -s" -trimpath -o speedtest main.go

FROM alpine:3.15
RUN apk add ca-certificates
WORKDIR /app
COPY --from=build_base /go/src/github.com/xiaoxinpro/speedtest-go-zh/speedtest .
COPY --from=build_base /go/src/github.com/xiaoxinpro/speedtest-go-zh/web/assets ./assets
COPY --from=build_base /go/src/github.com/xiaoxinpro/speedtest-go-zh/settings.toml .

EXPOSE 8989

CMD ["./speedtest"]
