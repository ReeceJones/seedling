FROM golang:alpine

RUN apk add build-base
RUN apk add helm

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go config.json ./

RUN go build -v -o /usr/local/bin/ ./...

EXPOSE 8081

ENV PATH $PATH:/usr/local/bin/seedling

CMD ["seedling", "-port", "8081"]