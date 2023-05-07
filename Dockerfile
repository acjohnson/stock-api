# Use the official Golang image as the base image
FROM golang:1.20.4-bullseye

# Set the working directory to /app
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go binary
RUN go build -o stock-api .

# Expose port 8080 for the API server
EXPOSE 8080

# Start the API server when the container starts
CMD ["./stock-api"]

