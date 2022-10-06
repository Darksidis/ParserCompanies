FROM golang:1.19-alpine as builder
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o main

RUN rm /app/main.go

COPY /actions/database.go ./

COPY /actions/sending_data.go ./
RUN go build -o /actions/sending_data
RUN rm /app/sending_data.go

COPY /actions/unpacking_archives.go ./
RUN go build -o /actions/unpacking_archives
RUN rm /app/unpacking_archives.go

COPY /actions/downloader_archives.go ./
RUN go build -o /actions/downloader_archives
RUN rm /app/downloader_archives.go

RUN rm /app/database.go

FROM alpine:3.16
WORKDIR /app

COPY --from=builder /app/main .
COPY actionsList.env ./
COPY db.env ./
COPY --from=builder actions /app/actions

RUN mkdir -p /app/storage
RUN mkdir -p /app/storage/archives
RUN mkdir -p /app/storage/files


CMD [ "/app/main" ]