FROM golang:1.19.1 as builder

LABEL org.opencontainers.image.description="Dockerized Karlsen Stratum Bridge"      
LABEL org.opencontainers.image.authors="onemorebsmith"  
LABEL org.opencontainers.image.source="https://github.com/onemorebsmith/karlsen-stratum-bridge"
              
WORKDIR /go/src/app
ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN go build -o /go/bin/app ./cmd/karlsenbridge


FROM gcr.io/distroless/base:nonroot
COPY --from=builder /go/bin/app /
COPY cmd/karlsenbridge/config.yaml /

WORKDIR /
ENTRYPOINT ["/app"]
