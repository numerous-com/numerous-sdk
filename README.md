Numerous Software Development Kit
=================================

💫 The python SDK for developing apps for the numerous platform.

📥 Simply install the SDK into your python environment with:

    pip install numerous

🛠 And then you can simply enter the following command, to get a list of possible
commands.

    numerous

👩🏼‍🎓 See the [numerous documentation](https://www.numerous.com/docs) for more information!

Badges
------

[![pypi badge](https://img.shields.io/pypi/v/numerous)](https://pypi.python.org/pypi/numerous)
[![Validate workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/validate.yml) 
[![Release workflow badge](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml/badge.svg)](https://github.com/numerous-com/numerous-sdk/actions/workflows/release.yml) 
![cli coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/cli-coverage.json)
![sdk coverage badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jfeodor/a9b9bfdfa0620696fba9e76223790f53/raw/sdk-coverage.json)

Development
===========

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

The numerous CLI enables tool development.

### Building and running

To build simply run `make build` without arguments, and the executable is stored
as `cli/build/numerous`

### Development

While developing you can run the CLI like below, inside the `cli` folder.

    go run .

    # e.g.
    go run . init
    go run . dev

From the root folder, you can lint with:

    make cli-lint

And you can run tests with

    make cli-test

### Trying out Numerous app engine development

In the `examples/numerous` folder are two apps `action.py` (containing
`ActionTool`), and `parameters.py` (containing `ParameterTool`). These can be
used to test the Numerous app engine development features.

**Note: You need an activate python environment with the python SDK installed.**
See the [python sdk development section](#development-of-python-sdk-) for
information about how to install it.

For example, if you built using `make cli-build`, you can run

```
./cli/build/numerous dev examples/numerous/parameters.py:ParameterApp
```
