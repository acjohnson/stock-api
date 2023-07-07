# Use the official Golang image as the base image
FROM golang:1.20.5-bookworm

# Set the working directory to /app
WORKDIR /app

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y curl wget gnupg2 ca-certificates chromium

# Copy the source code into the container
COPY . .

# Build the Go binary
RUN go build -o stock-api .

# Expose port 8080 for the API server
EXPOSE 8080

# Start the API server when the container starts
CMD ["./stock-api"]
