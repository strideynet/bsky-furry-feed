## Build layer
FROM golang:1.21.0-bullseye AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -o /app/bffsrv ./cmd/bffsrv

## Deploy layer
FROM gcr.io/distroless/base-debian11

COPY --from=build /app/bffsrv /app/bffsrv

EXPOSE 1337
EXPOSE 1338

USER nonroot:nonroot

ENTRYPOINT ["/app/bffsrv"]