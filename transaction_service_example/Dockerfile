FROM golang:1.21.5 as builder
WORKDIR /builder

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./transaction ./main.go

FROM scratch as prod
WORKDIR /prod

COPY .env .
COPY --from=builder /builder/transaction .

CMD [ "./transaction" ]