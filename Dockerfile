# Use an existing docker image as a base
FROM golang:alpine AS build

# Create a directory in the container
WORKDIR /src

# Copy everything from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main cmd/main.go

# Run stage
FROM gcr.io/distroless/base-debian10

WORKDIR /app

# Copy the binary from the build stage to the final stage
COPY --from=build /src/main /app/main

# Attempt to execute a file called main
CMD [ "./main" ]