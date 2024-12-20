# Contributing to Numerous SDK

Here are the guidelines for contributing to Numerous SDK! They are chosen with
the aim of creating a good environment for developing Numerous SDK to a high
level of quality and consistency.

## Linting, pre-commit and testing

We recommend setting up the pre-commit hooks defined in the repository
to ensure that all linters and tests are passing before a commit or push.

Reviews in general require tests and linters to pass.

## Code style and conventions

We expect contributors to familiarize themselves with the
[Google Go Style Guide](https://google.github.io/styleguide/go/), in particular
the [actual style guide itself](https://google.github.io/styleguide/go/style),
and the set of
[best practices](https://google.github.io/styleguide/go/best-practices).

These documents can be referenced in reviews for code style suggestions.

Above all, we encourage developers to prioritize designing their code so that it
is readable and understandable by other developers (or themselves, in the
future!), easy to modify, and testable.

We also expect new features added to have an associated test suite, which
should be usable as documentation to understand the expectations of the new
code.

## Documentation

In the python SDK, we document modules, functions, classes, and methods using
docstrings. We use the [Google Docstring format](https://google.github.io/styleguide/pyguide.html#s3.8-comments-and-docstrings).

## Git workflow

When creating a pull request with changes, please consider the following
points:

 * We use
   [conventional commits](https://www.conventionalcommits.org/en/v1.0.0),
   for automated semantic versioning.
 * Make sure the pull request branch is rebased on top of `main`.
 * Even within your pull request branch, try to ensure that linters and tests
   pass locally, before pushing.
