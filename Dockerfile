FROM golang:1.9 as builder
COPY . /
WORKDIR /
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch
COPY --from=builder /main /main
CMD ["/main"]
