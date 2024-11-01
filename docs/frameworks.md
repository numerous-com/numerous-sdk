# Frameworks

The Numerous SDK supports a number of popular Python web frameworks.

The `frameworks` package provides framework-specific implementations of Numerous features. Currently, the main framework-specific function is `get_session()`, which handles user authentication by accessing session information through cookies in a way that's compatible with each framework's request handling.

## Dash

Plotly Dash lets you build interactive data dashboards using Python. Common uses include:
- Creating charts and graphs that users can filter and explore
- Building business dashboards to monitor KPIs and metrics
- Sharing data analysis results with non-technical team members
- Developing internal tools for data exploration

No JavaScript or HTML knowledge is needed - everything can be built in Python. Dash applications can be easily shared via web browsers, making them perfect for teams that need to collaborate around data.

### ::: numerous.frameworks.dash.get_session
    options:
        show_root_heading: true

###
## FastAPI

FastAPI is a modern web framework for building APIs with Python. It's known for being:
- Fast to code: Write APIs with less code and fewer bugs
- Easy to use: Clear error messages and built-in documentation
- Production-ready: Used by companies of all sizes
- Automatic API documentation: Interactive docs are created automatically

Common uses include:
- Building backend services for web and mobile apps
- Creating microservices that need to handle many requests
- Developing data APIs that connect to databases
- Setting up internal tools and services

### ::: numerous.frameworks.fastapi.get_session
    options:
        show_root_heading: true

###
## Flask

Flask is a lightweight and flexible web framework that's perfect for both small projects and large applications. Key features include:
- Simple to learn and use: Great for beginners
- Highly customizable: Add only what you need
- Large ecosystem of extensions
- Works well for both websites and APIs

Common uses include:
- Building web applications of any size
- Creating REST APIs
- Prototyping new ideas quickly
- Adding web interfaces to existing Python projects

### ::: numerous.frameworks.flask.get_session
    options:
        show_root_heading: true

###
## Marimo

Marimo is a modern Python notebook that turns your data analysis into interactive web apps. Features include:
- Interactive elements that update in real-time
- Clean, reproducible code execution
- Easy sharing of notebooks as web apps
- Better version control than traditional notebooks

Perfect for:
- Data analysis and exploration
- Creating interactive reports
- Teaching and learning Python
- Sharing research findings

### ::: numerous.frameworks.marimo.get_session
    options:
        show_root_heading: true

###
## Panel

Panel is a powerful framework for creating custom data dashboards and apps. Highlights include:
- Works with multiple plotting libraries (Plotly, Bokeh, Matplotlib)
- Easy to convert existing visualizations into interactive dashboards
- Supports both simple and complex layouts
- Can be used with Jupyter notebooks or as standalone apps

Ideal for:
- Data scientists who need to share interactive visualizations
- Creating complex dashboards with multiple data sources
- Building data-driven web applications
- Turning analysis notebooks into deployed applications

### ::: numerous.frameworks.panel.get_session
    options:
        show_root_heading: true

###
## Streamlit

Streamlit turns data scripts into shareable web apps in minutes. Key benefits include:
- Extremely easy to learn and use
- Fast development of data apps
- Built-in widgets for interaction
- Simple deployment process

Popular uses include:
- Creating data science demos
- Building machine learning model interfaces
- Sharing data insights with stakeholders
- Rapid prototyping of data applications

### ::: numerous.frameworks.streamlit.get_session
    options:
        show_root_heading: true
