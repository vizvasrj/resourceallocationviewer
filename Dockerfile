# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

EXPOSE 8080
# Set the command to run the executable when the container starts
CMD ["./main"]