# Get Golang 1.19.4
FROM golang:1.19.4-bullseye

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies.
RUN go mod download

# Copy the source from the current directory
# to the Working Directory inside the container
COPY . .

# Build the Go app
# as the executable file
RUN go build -o team ./cmd

# Expose port
EXPOSE 4001

# Run Application
CMD ["./team"]