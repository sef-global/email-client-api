FROM golang:1.22
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
    SMTPPASS= \
    URL= 

CMD ["./app"]
