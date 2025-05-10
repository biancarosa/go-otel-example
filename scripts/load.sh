#!/bin/bash

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo "Vegeta is not installed. Installing..."
    go install github.com/tsenart/vegeta/v12@latest
    
    # Add GOPATH/bin to PATH if it's not already there
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
    
    # Verify installation
    if ! command -v vegeta &> /dev/null; then
        echo "Failed to install Vegeta. Please install it manually:"
        echo "go install github.com/tsenart/vegeta/v12@latest"
        echo "Then add $HOME/go/bin to your PATH"
        exit 1
    fi
fi

# Default values
DURATION="30s"
RATE="100"
TARGETS_FILE="scripts/targets.txt"
REPORT_FILE="load-test-report.html"

# Create targets file if it doesn't exist
mkdir -p scripts
cat > $TARGETS_FILE << EOL
GET http://localhost:8080/
GET http://localhost:8080/user
GET http://localhost:8080/health
GET http://localhost:8080/metrics
EOL

echo "Starting load test with the following parameters:"
echo "Duration: $DURATION"
echo "Rate: $RATE requests per second"
echo "Targets file: $TARGETS_FILE"

# Run the load test
vegeta attack \
    -duration=$DURATION \
    -rate=$RATE \
    -targets=$TARGETS_FILE \
    | vegeta report \
    | tee load-test-results.txt

# Generate HTML report
vegeta attack \
    -duration=$DURATION \
    -rate=$RATE \
    -targets=$TARGETS_FILE \
    | vegeta report -type=html > $REPORT_FILE

echo "Load test completed!"
echo "Results saved to load-test-results.txt"
echo "HTML report saved to $REPORT_FILE" 