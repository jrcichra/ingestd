FROM golang:1.19.0-alpine3.16
WORKDIR /ingestd
RUN apk add make git
COPY . . 
RUN make

FROM alpine:3.16
WORKDIR /ingestd
EXPOSE 8080
COPY --from=0 /ingestd/ingestd .
CMD ./ingestd