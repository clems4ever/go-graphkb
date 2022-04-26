FROM golang:1.18 AS go-builder

WORKDIR /go/src/
COPY go.mod go.sum ./

RUN go mod download

COPY cmd cmd
COPY graphkb graphkb
COPY internal internal
RUN cd cmd/go-graphkb && GOOS=linux GOARCH=amd64 go build -o go-graphkb main.go
RUN cd cmd/go-graphkb && GOOS=linux GOARCH=amd64 go build -o datasource-csv main.go



FROM node:16-alpine AS node-builder

WORKDIR /node/src/

COPY web .

RUN yarn install && yarn build

COPY --from=go-builder /go/src/cmd/go-graphkb/go-graphkb ./
COPY --from=go-builder /go/src/cmd/go-graphkb/datasource-csv ./
