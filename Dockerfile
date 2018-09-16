FROM golang:1.10-alpine as builder
WORKDIR /go/src/github.com/jfrog/jfrog-cli-go
COPY . /go/src/github.com/jfrog/jfrog-cli-go
RUN CGO_ENABLED=0 GOOS=linux go build github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/jfrog

FROM alpine:3.7
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder /go/src/github.com/AlexeiVainshtein/jfrog-cli-gojfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog

ENTRYPOINT [ "/usr/local/bin/jfrog" ]
CMD ["--help"]
