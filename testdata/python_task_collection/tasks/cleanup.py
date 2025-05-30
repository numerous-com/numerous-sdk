from numerous import task

@task
def cleanup_data(data: dict) -> str:
    """Clean up processed data"""
    # Simulate cleanup logic
    return f"Cleaned {len(data)} entries" 