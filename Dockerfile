FROM golang:1.17.7-alpine3.15
WORKDIR /ingestd
RUN apk add make
COPY . . 
RUN make

FROM alpine:3.15
WORKDIR /ingestd
EXPOSE 8080
COPY --from=0 /ingestd/ingestd .
CMD ./ingestd