# Numerous Software Development Kit

[![pypi badge](https://img.shields.io/pypi/v/numerous)](https://pypi.python.org/pypi/numerous)
[![Validate workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml)
[![Release workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml)
![cli coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/cli-coverage.json)
![sdk coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/sdk-coverage.json)

Welcome to the Numerous SDK documentation!

This is the documentation for the SDK, if you are looking an overview of the
Numerous platform, please visit the
[Numerous platform documentation](https://numerous.com/docs).

📥 To begin, install the SDK into your Python environment with:

    pip install numerous

🛠 The installation includes two main components: the CLI and the Numerous
Python package.

## Command Line Interface (CLI) - a tool for managing your apps

Use the CLI to deploy and manage apps on the Numerous platform:

- `numerous init` - Create a new Numerous app
- `numerous deploy` - Deploy your app to production

## Python SDK package - integrate Numerous features into Python-based web apps

The SDK enables you to interact with Numerous services programmatically in
your Python web applications.

- Store and organize data (JSON documents, and files) with collections.
- Access information about users and manage user interaction with sessions.

# SDK Development

This section contains information about how to develop the SDK itself for
developers interested in contributing to the SDK.

Most common tasks are defined in the `Makefile`. Use `make help` to get an
overview.

In order to setup pre-commit hooks, use [pre-commit](https://pre-commit.com/) to
to setup hooks for linters and tests. This requires pre-commit to be installed
of course, and it is included in the python SDK development dependencies.

To install pre-commit and pre-push hooks:

    pre-commit install

And you can run them on demand:

    pre-commit run --all

## Development of Python SDK 🐍

Create a virtual environment and activate it:

    python -m venv ./venv
    ./venv/bin/activate

Install the package in editable mode (including development dependencies):

    pip install -e ./python[dev]

Run the tests:

    make sdk-test

And the linters:

    make sdk-lint

## Development of Go CLI 🐹

To build, run `make cli-build`, and the executable is stored as `build/numerous`

While developing, you can run the CLI like below:

    # Run the CLI
    go run .

    # e.g.
    go run . init
    go run . dev

You can lint with:

    make cli-lint

And you can run tests with:

    make cli-test
