FROM golang:alpine AS binarybuilder
WORKDIR /file-vault-lite/
COPY . .
RUN go build -o app main.go

FROM alpine:latest
RUN echo http://dl-2.alpinelinux.org/alpine/edge/community/ >>/etc/apk/repositories && apk --no-cache --no-progress add \
  tzdata \
  ca-certificates
WORKDIR /file-vault-lite/
COPY --from=binarybuilder /file-vault-lite/app .
VOLUME ["/file-vault-lite/uploads"]
EXPOSE 8080
CMD ["/file-vault-lite/app"]
