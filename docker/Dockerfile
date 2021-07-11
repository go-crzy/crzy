FROM golang:1.16-alpine as builder
RUN apk add --no-cache wget git

ARG SOURCE_COMMIT
ARG DOCKER_TAG
ENV SOURCE_COMMIT $SOURCE_COMMIT
ENV DOCKER_TAG $DOCKER_TAG

WORKDIR /build
COPY . .
RUN go mod download && CGO_ENABLED=0 go build \
  -ldflags "-s -w -X github.com/go-crzy/crzy/pkg.version=${DOCKER_TAG:-dev} -X github.com/go-crzy/crzy/pkg.commit=${SOURCE_COMMIT:-unknown}" \
  -o crzy .

FROM golang:1.16-alpine
RUN apk add --no-cache wget git gcc musl-dev
WORKDIR /app
COPY --from=builder /build/crzy /bin/crzy
ENV REPOSITORY myrepo
EXPOSE 8080
EXPOSE 8081
ENTRYPOINT ["/bin/crzy"]
