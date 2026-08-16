FROM golang:1.21-alpine
ENV GOTOOLCHAIN=local
WORKDIR /app
COPY . .
RUN go build -o /app/bin .
CMD ["sh"]
