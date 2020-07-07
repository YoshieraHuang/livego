# FROM golang:latest as builder
# WORKDIR /app
# ENV GOPROXY https://goproxy.io
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o livego -v

ARG RTMP_PORT=1935
ARG HTTP_FLV_PORT=7001

FROM alpine:latest
RUN mkdir -p /app
WORKDIR /app
# COPY --from=builder /app/livego .
COPY livego /app/
EXPOSE ${RTMP_PORT} ${HTTP_FLV_PORT}
ENTRYPOINT ["./livego"]
