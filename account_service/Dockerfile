FROM golang:1.21.5 as builder
WORKDIR /builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./account ./cmd/account_service/main.go

FROM scratch as prod
WORKDIR /prod

COPY .env .
COPY --from=builder /builder/account .

EXPOSE 9999
CMD [ "./account" ]