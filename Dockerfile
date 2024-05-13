# syntax=docker/dockerfile:1
FROM golang:1.22

WORKDIR /app

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/ESS-excel-reader ./app/

RUN #go test -v ./...

#RUN addgroup -g 1000 appgroup
#RUN adduser -D -u 1000 appuser -G appgroup

#USER appuser

ENTRYPOINT ["./build/ESS-excel-reader"]