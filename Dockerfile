FROM golang:1.23-alpine
   
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o camera-monitor .

CMD ["./camera-monitor"]