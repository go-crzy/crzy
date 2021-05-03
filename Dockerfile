FROM golang:1.16-alpine as builder
RUN apk add --no-cache wget git
WORKDIR /build
COPY . . 
RUN go mod download && CGO_ENABLED=0 go build -o crzy .

FROM golang:1.16-alpine
RUN apk add --no-cache wget git
WORKDIR /app
COPY --from=builder /build/crzy /bin/crzy
EXPOSE 8080
EXPOSE 8081
ENTRYPOINT ["/bin/crzy"]
CMD ["-server"]
