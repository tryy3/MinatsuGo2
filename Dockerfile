# Use an existing docker image as a base
FROM ubuntu:latest

# Create a directory in the container
WORKDIR /app

# Link a folder from the host system to the container
VOLUME [ "/app" ]

# Attempt to execute a file called main
CMD [ "./main" ]