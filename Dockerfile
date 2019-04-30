FROM golang:1.12 as builder
COPY helloworld.go .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o helloworld

FROM scratch
COPY --from=builder /go/helloworld /helloworld
CMD ["/helloworld"]
