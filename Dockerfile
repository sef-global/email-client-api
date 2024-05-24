FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app ./cmd/api

EXPOSE 4000

ENV DB_DSN= \
    SMTPPORT= \
    SMTPSENDER= \
    SMTPHOST= \
    SMTPUSERNAME= \
    SMTPPASS=

CMD ["./app"]