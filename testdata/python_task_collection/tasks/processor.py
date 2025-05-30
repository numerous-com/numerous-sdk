from numerous import task
import pandas as pd

@task
def data_processor(data: dict) -> dict:
    """Process data using pandas"""
    df = pd.DataFrame(data)
    processed = df.fillna(0)
    return processed.to_dict() 