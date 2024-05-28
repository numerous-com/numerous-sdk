# CHANGELOG



## v0.3.2 (2024-05-28)

### Build

* build(deps): bump golang.org/x/net from 0.22.0 to 0.23.0 in /cli (#3)

Bumps [golang.org/x/net](https://github.com/golang/net) from 0.22.0 to 0.23.0.
- [Commits](https://github.com/golang/net/compare/v0.22.0...v0.23.0)

---
updated-dependencies:
- dependency-name: golang.org/x/net
  dependency-type: indirect
...

Signed-off-by: dependabot[bot] &lt;support@github.com&gt;
Co-authored-by: dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt; ([`c2472bf`](https://github.com/numerous-com/numerous-sdk/commit/c2472bf3f55000150f7e9f546607bde8ca991579))

* build(deps): bump github.com/lestrrat-go/jwx in /cli (#1)

Bumps [github.com/lestrrat-go/jwx](https://github.com/lestrrat-go/jwx) from 1.2.28 to 1.2.29.
- [Release notes](https://github.com/lestrrat-go/jwx/releases)
- [Changelog](https://github.com/lestrrat-go/jwx/blob/v1.2.29/Changes)
- [Commits](https://github.com/lestrrat-go/jwx/compare/v1.2.28...v1.2.29)

---
updated-dependencies:
- dependency-name: github.com/lestrrat-go/jwx
  dependency-type: direct:production
...

Signed-off-by: dependabot[bot] &lt;support@github.com&gt;
Co-authored-by: dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt; ([`6b9c01a`](https://github.com/numerous-com/numerous-sdk/commit/6b9c01a21fc857c66e13143cd9def4b685dc643e))

### Documentation

* docs: typo and phrasing in CONTRIBUTING.md ([`079c661`](https://github.com/numerous-com/numerous-sdk/commit/079c661a09a367253f20a55d9a595aadaa0ca30b))

* docs: add CONTRIBUTING.md ([`e71c27b`](https://github.com/numerous-com/numerous-sdk/commit/e71c27b0dd69211eb4e916793fd09f4a4d26e755))

### Fix

* fix(cli): bug causing logs to not be correctly read ([`e66e051`](https://github.com/numerous-com/numerous-sdk/commit/e66e051567ecc4ddffd96af4aee1798130869426))


## v0.3.1 (2024-05-08)

### Fix

* fix(cli): add `.env` to default excluded files ([`0bf8a32`](https://github.com/numerous-com/numerous-sdk/commit/0bf8a322d6176adb8d53d028e2ed786fc1f7c43a))

* fix(cli): improve error output for `push` ([`2b05a1f`](https://github.com/numerous-com/numerous-sdk/commit/2b05a1f5aec09b53539e6494260d7291cb73ea15))

* fix(cli): add `.env` file to &#39;.gitignore&#39; ([`098f8f5`](https://github.com/numerous-com/numerous-sdk/commit/098f8f50760c453505b405383557236d1dc544a6))


## v0.3.0 (2024-05-07)

### Feature

* feat(cli): `push` reads `.env` and sends secrets

Read and parse `.env` in the app directory, and send the parsed
environment to the server for it to configure the resulting app
deployment with the secrets from the `.env`. ([`c019861`](https://github.com/numerous-com/numerous-sdk/commit/c019861e47d99b2109e0091cd57f29e0c950fd72))


## v0.2.0 (2024-05-07)

### Feature

* feat(cli): create `.app_id.txt` to store App ID

Includes backwards compatibility for projects initialized with `.tool_id.txt`. Also includes various improvements to output
formatting related to files bootstrapping. ([`7a8f682`](https://github.com/numerous-com/numerous-sdk/commit/7a8f68238319e714c94851b43904f0e5a3e8f703))

### Fix

* fix(cli): improved output for commands, and minor refactors

Use common error printing functions, and improve the phrasing of some error and informative messages. ([`15bdc92`](https://github.com/numerous-com/numerous-sdk/commit/15bdc92e8747cf40e4a1ebe3139c4a317ae8b784))

* fix(cli): bug `numerous push` fails due to only reading deprecated App ID file

Use the common function read the App ID, which can read from `.app_id.txt`, since it is now created by the `numerous init`command ([`f80458b`](https://github.com/numerous-com/numerous-sdk/commit/f80458b3a49aa8e2e1823d82c3e8a02274214695))


## v0.1.3 (2024-05-06)

### Fix

* fix(cli): login and logout output ([`becb0ae`](https://github.com/numerous-com/numerous-sdk/commit/becb0ae7145f4a46a3b682be961f70ef681029cf))


## v0.1.2 (2024-05-06)

### Fix

* fix(cli): improve &#39;numerous log&#39; output

* A bit of refactoring
* Adds flag to print timestamps
* Adds standardized error printing functions ([`1b3e7e5`](https://github.com/numerous-com/numerous-sdk/commit/1b3e7e52545d290b330eba6690a0191fabc1df14))


## v0.1.1 (2024-05-06)

### Documentation

* docs: references to examples in README.md ([`e12d309`](https://github.com/numerous-com/numerous-sdk/commit/e12d30997ce6683708dfa1af326d625cc9583119))

### Fix

* fix(cli): improve &#39;init&#39; output, and use &#39;.app_id.txt&#39;

* Make output of the &#39;numerous init&#39; command more readable, friendly and colorful.
* Read &#39;.app_id.txt&#39; for the App ID, falling back to the old &#39;tool_id.txt&#39;.
* Added a test helper for writing to a file.
* Removed a println in the bootstrap code.
* Fixed some print statements that were lower cased. ([`75798d1`](https://github.com/numerous-com/numerous-sdk/commit/75798d1827eeba9dc797057e05db867523d78af7))


## v0.1.0 (2024-05-03)
