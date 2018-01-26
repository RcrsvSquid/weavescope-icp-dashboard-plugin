FROM golang:1.9 as builder
COPY . /
WORKDIR /
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine
COPY --from=builder /main /main
COPY test.json /test.json
CMD ["/main"]
