# syntax=docker/dockerfile:1
FROM golang:1.22

WORKDIR /app

COPY app/go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./app/ESS-excel-reader ./app/main.go

RUN #go test -v ./...

#RUN addgroup -g 1000 appgroup
#RUN adduser -D -u 1000 appuser -G appgroup

#USER appuser

ENTRYPOINT ["./app/ESS-excel-reader"]