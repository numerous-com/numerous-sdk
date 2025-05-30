# Docker validator task collection
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    jq \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY tasks/ ./tasks/
COPY app.py .

# Make shell scripts executable
RUN chmod +x tasks/entrypoint.sh

# Expose port for API endpoints
EXPOSE 8080

# Default command runs the web server
CMD ["python", "app.py"] 