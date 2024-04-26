FROM golang:latest

WORKDIR /go/src/app

# COPY go.* .

# RUN go mod download


COPY . .

# RUN go build -o myapp

CMD ["./myapp"]