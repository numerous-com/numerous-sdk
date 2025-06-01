FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy all source code
COPY . .

# Make task runner executable
RUN chmod +x task_runner.py

RUN echo "Building validator python tasks"

# Default entrypoint - can be overridden with specific task
ENTRYPOINT ["python", "task_runner.py"]

# Default command - runs validate_environment if no task specified
CMD ["validate_environment"] 