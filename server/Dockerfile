ARG GO_VERSION=1.23.0
FROM golang:${GO_VERSION}

WORKDIR /app
COPY . .

RUN go build -o server

EXPOSE 8080

CMD [ "./server" ]
