FROM golang:alpine as builder
RUN apk add git
RUN mkdir /build
COPY . /build/
WORKDIR /build
RUN go build -o invertergui ./cmd/invertergui
FROM alpine
RUN adduser -S -D -H -h /app inverteruser
RUN addgroup inverteruser dialout
USER inverteruser
COPY --from=builder /build/invertergui /app/
WORKDIR /app
ENTRYPOINT [ "./invertergui" ]
CMD []