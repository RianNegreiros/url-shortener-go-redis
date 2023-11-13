FROM golang:1.21

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o /api

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /api /api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/api"]