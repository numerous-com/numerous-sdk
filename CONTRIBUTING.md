# Contributing to Numerous SDK

Here are the guidelines for contributing to Numerous SDK! They are chosen with
the goal of creating a good environment for developing Numerous SDK to a high
level of quality and consistency.

## Linting, pre-commit and testing

We recommend setting up the pre-commit hooks defined in the repositoriy, in
to make sure that linters and tests all pass, before committing and pushing.

Reviews in general require tests and linters to pass.

## Code style and conventions

We expect contributors to familiarize themselves with the
[Google Go Style Guide](https://google.github.io/styleguide/go/), in particular
the [actual style guide itself](https://google.github.io/styleguide/go/style),
and the set of
[best practices](https://google.github.io/styleguide/go/best-practices).

These documents may be referenced in reviews for code style suggestions.

Above all, we encourage developers to prioritize designing their code so that it
is readable and understandable for other developers (or themselves, in the
future!), easy to modify, and testable.

We also expect new features added to have an assoiciated test suite, which
should be usable as documentation to understand the expectations of the new
code.

## Git workflow

When you create a pull request with changes, please consider the following
points:

 * We use
   [conventional commits](https://www.conventionalcommits.org/en/v1.0.0),
   for automated semantic versioning.
 * Make sure the pull request branch is rebased on top of `main`.
 * Even within your pull request branch, try to ensure that linters and tests
   pass locally, before pushing.
