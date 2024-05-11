FROM golang:1.22.3-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd
RUN go build -o main .

# EXPOSE 443
# EXPOSE 80

CMD ["./main"]
