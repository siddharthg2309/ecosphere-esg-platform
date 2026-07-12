FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ecosphere ./cmd/api

FROM alpine:3.21
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/ecosphere /usr/local/bin/ecosphere
EXPOSE 8080
ENTRYPOINT ["ecosphere"]
