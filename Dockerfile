# Building Backend
FROM golang:alpine as wallet-server

WORKDIR /source
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs -o /dist ./pkg/main.go

# Runtime
FROM golang:alpine

COPY --from=wallet-server /dist /wallet/server

EXPOSE 8445

CMD ["/wallet/server"]
