FROM registry.access.redhat.com/ubi9/go-toolset:1.24.6-1758501173 as build
WORKDIR /app
COPY cmd ./cmd
COPY go.mod ./go.mod
COPY go.sum ./go.sum
USER 0
RUN go build ./cmd/operator-registry-proxy/main.go

FROM quay.io/flightctl/flightctl-base:9.6-1758714456
COPY --from=build /app/main /app/main
COPY examples /app/examples
WORKDIR /app
EXPOSE 8080
CMD ./main