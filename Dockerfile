FROM golang:1.23-bookworm

# Install libpcap for packet capture
RUN apt-get update && apt-get install -y \
    libpcap-dev \
    lsof \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Allow Go to download newer toolchain if needed
ENV GOTOOLCHAIN=auto

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN go build -o bytefall ./cmd/bytefall

# Run in demo mode by default (no root needed)
CMD ["./bytefall", "-demo"]
