RUN apt update && apt install ghdl -y
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server .
ENTRYPOINT ["/app/server"]