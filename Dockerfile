FROM golang:1.16-alpine3.14
WORKDIR /ingestd
RUN apk add make
COPY . . 
RUN make

FROM alpine:3.14
WORKDIR /ingestd
EXPOSE 8080
COPY --from=0 /ingestd/ingestd .
CMD ./ingestd