## Build layer
FROM golang:1.19.3-buster AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN go build -o /bffsrv ./cmd/bffsrv

## Deploy layer
FROM gcr.io/distroless/base-debian10

WORKDIR /
COPY --from=build /bffsrv /bffsrv

EXPOSE 1337
EXPOSE 1338

USER nonroot:nonroot

ENTRYPOINT ["/bffsrv"]