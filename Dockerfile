FROM golang:1.23.3-bullseye

WORKDIR /app

COPY go.mod ./
RUN go mod download all

COPY . ./
RUN go build chord

CMD [ "./chord" ]
