#!/bin/bash

# Set environment variables for local development

#export NUMEROUS_GRAPHQL_HTTP_URL="https://api.numerous.com/query"
#export NUMEROUS_GRAPHQL_WS_URL="wss://api.numerous.com/query"

#export NUMEROUS_GRAPHQL_HTTP_URL="https://api.numerous-staging.site/query"
#export NUMEROUS_GRAPHQL_WS_URL="wss://api.numerous-staging.site/query"

export NUMEROUS_GRAPHQL_HTTP_URL="http://localhost:8080/query"
export NUMEROUS_GRAPHQL_WS_URL="ws://localhost:8080/query"

# Explicitly enable file-based credentials storage instead of keyring
# This solves the keyring access issues in WSL
export NUMEROUS_LOGIN_USE_KEYRING=false

# Enable debug logging (optional)
export NUMEROUS_LOG_LEVEL=debug

echo "=== Numerous CLI WSL Helper ==="
echo "GraphQL HTTP URL: $NUMEROUS_GRAPHQL_HTTP_URL"
echo "GraphQL WS URL: $NUMEROUS_GRAPHQL_WS_URL"
echo "Using file-based credentials storage (~/.numerous/credentials.json)"
echo "=============================="

# Run the CLI with arguments
cd "$(dirname "$0")"
go run main.go "$@"

# Print helpful message for login
if [[ "$*" == *"login"* ]]; then
  echo ""
  echo "Note: Your credentials will be stored in ~/.numerous/credentials.json"
  echo "To check login status, run: ./run-local-wsl.sh status"
fi 