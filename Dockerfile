FROM golang:1.16-alpine as builder

RUN mkdir /build
COPY . /build/
WORKDIR /build
RUN CGO_ENABLED=0 go build -o invertergui ./cmd/invertergui

FROM scratch

# Group ID 20 is dialout, needed for tty read/write access
USER 3000:20
COPY --from=builder /build/invertergui /bin/
ENTRYPOINT [ "/bin/invertergui" ]
EXPOSE 8080
