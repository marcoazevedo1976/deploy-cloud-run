FROM golang:1.24 AS build

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cloudrun

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/cloudrun .

EXPOSE 8080
CMD ["./cloudrun"]