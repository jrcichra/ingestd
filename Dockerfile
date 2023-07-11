FROM golang:1.20.6-bullseye
WORKDIR /ingestd
COPY . . 
RUN CGO_ENABLED=0 go build -o ingestd

FROM gcr.io/distroless/static-debian11
WORKDIR /
EXPOSE 8080
COPY --from=0 /ingestd/ingestd .
ENTRYPOINT [ "/ingestd" ]
