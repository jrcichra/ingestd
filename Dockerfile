FROM golang:1.21.6-bullseye as builder
WORKDIR /ingestd
COPY . . 
RUN CGO_ENABLED=0 go build -o ingestd

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
EXPOSE 8080
COPY --from=builder /ingestd/ingestd .
ENTRYPOINT [ "/ingestd" ]
