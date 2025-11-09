FROM golang:1.25 AS build
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM gcr.io/distroless/static-debian12
COPY --from=build /server /server
EXPOSE 8000
CMD ["/server"]