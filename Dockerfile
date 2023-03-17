# Use the official golang image as the base image
FROM golang:1.20.2 as builder

# Set the working directory in the container
WORKDIR /app

# Copy the rest of the application files to the working directory
COPY . .

# Download the dependencies
RUN go mod download

# Build the application
RUN go build -o main .

# Use the light weight alpine image as the base image
FROM alpine:3.17.2

# Set the working directory in the container
WORKDIR /app

# Copy the built application from the builder image
COPY --from=builder /app/main .

# Expose port 3000 to the host
EXPOSE 3000

# Run the application
CMD ["./main"]