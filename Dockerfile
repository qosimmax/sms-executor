# Build the project
FROM golang:1.20 as builder

WORKDIR /go/src/github.com/qosimmax/sms-executor
ADD . .

RUN make build
#RUN make test

# Create production image for application with needed files
FROM golang:1.20.5-alpine3.18

EXPOSE 8000

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/qosimmax/sms-executor .

CMD ["./bin/sms-executor"]
