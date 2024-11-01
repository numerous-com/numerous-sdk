Numerous Software Development Kit
=================================

Welcome to the Numerous SDK documentation!

This is the documentation for the SDK, if you are looking an overview of the Numerous platform, please visit the [Numerous platform documentation](https://numerous.com/docs).

📥 To begin, simply install the SDK into your python environment with:

    pip install numerous

🛠 The installation includes two main components; the CLI and the numerous Python package.

**Command Line Interface (CLI)** - A tool for managing Numerous apps on the platform:

   Use the CLI to deploy and manage apps on the Numerous platform:

   - `numerous init` - Create a new Numerous app

   - `numerous deploy` - Deploy your app to production

**The numerous Python package** - A Python package for integrating Numerous features in Python-based web apps:

   The SDK enables you to interact with Numerous services programmatically in your Python web applications, such as collections for storing data, sessions for managing user interactions, and users for identifying and managing users.


**Bagdes:**

[![pypi badge](https://img.shields.io/pypi/v/numerous)](https://pypi.python.org/pypi/numerous)
[![Validate workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml) 
[![Release workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml) 
![cli coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/cli-coverage.json)
![sdk coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/sdk-coverage.json)

SDK Development
===============

This section contains information about how to develop the SDK itself for developers interested in contributing to the SDK.

Most common tasks are defined in the `Makefile`. Use `make help` to get an
overview.

In order to setup pre-commit hooks, use [pre-commit](https://pre-commit.com/) to
to setup hooks for linters and tests. This requires pre-commit to be installed
of course, and it is included in the python SDK development dependencies.

To install pre-commit and pre-push hooks

    pre-commit install

And you can run them on demand

    pre-commit run --all

Development of python SDK 🐍
----------------------------

Create a virtual environment and activate it

    python -m venv ./venv
    ./venv/bin/activate

Install the package in editable mode (including development dependencies)

    pip install -e ./python[dev]

Run the tests

    make sdk-test

And the linters

    make sdk-lint

Development of go CLI 🐹
------------------------

The numerous CLI enables app development.

### Building and running

To build simply run `make cli-build`, and the executable is stored
as `build/numerous`

### Development

While developing you can run the CLI like below.

    # Run the CLI
    go run .

    # e.g.
    go run . init
    go run . dev

You can lint with:

    make cli-lint

And you can run tests with

    make cli-test
