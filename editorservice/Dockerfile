FROM golang:1.14 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pb/ pb/

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o editorservice .

FROM scratch

WORKDIR /app

COPY --from=builder /app/editorservice /app/editorservice

ENTRYPOINT ["/app/editorservice"]

