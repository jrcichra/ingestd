FROM golang:1.15.2-alpine3.12
WORKDIR /ingestd
RUN apk add make
COPY . . 
RUN make

FROM alpine:3.12.0
WORKDIR /ingestd
COPY --from=0 /ingestd/ingestd .
CMD /ingestd/ingestd