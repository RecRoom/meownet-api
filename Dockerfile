FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/slop-net .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /app/slop-net .
COPY data/ ./data/
COPY db/seeds/ ./db/seeds/
COPY db/words/ ./db/words/

EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["./slop-net"]
