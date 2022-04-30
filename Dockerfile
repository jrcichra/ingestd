FROM golang:1.18.1-alpine3.15
WORKDIR /ingestd
RUN apk add make git
COPY . . 
RUN make

FROM alpine:3.15
WORKDIR /ingestd
EXPOSE 8080
COPY --from=0 /ingestd/ingestd .
CMD ./ingestd