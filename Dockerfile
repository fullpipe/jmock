# STEP 1 build executable binary
FROM golang AS builder

ENV GO111MODULE=on
WORKDIR /app

COPY . .

#Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s"

# STEP 2 build a small image
FROM scratch

# Copy our static executable.
COPY --from=builder /app/jmock /go/bin/jmock

# Expose port 9090
EXPOSE 9090

# Run the hello binary.
ENTRYPOINT ["/go/bin/jmock"]

# By default lookup for mocks in /mocks directory
CMD ["/mocks/*.json", "--port", "9090"]
