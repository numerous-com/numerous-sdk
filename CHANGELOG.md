# CHANGELOG


## v0.6.0 (2024-06-14)

### Feature

* feat(cli): app subcommand (#11) ([`351680e`](https://github.com/numerous-com/numerous-sdk/commit/351680ec009253f0030134cbf2cdf67bc3bc1a31))




## v0.5.0 (2024-06-14)

### Build

* build(deps): bump github.com/vektah/gqlparser/v2 in /cli (#10) ([`5d2d9f0`](https://github.com/numerous-com/numerous-sdk/commit/5d2d9f0519f8346317d513ad4084d03ad2ff4f08))

  > 
  > Bumps [github.com/vektah/gqlparser/v2](https://github.com/vektah/gqlparser) from 2.5.11 to 2.5.15.
  > - [Release notes](https://github.com/vektah/gqlparser/releases)
  > - [Commits](https://github.com/vektah/gqlparser/compare/v2.5.11...v2.5.15)
  > 
  > ---
  > updated-dependencies:
  > - dependency-name: github.com/vektah/gqlparser/v2
  >   dependency-type: direct:production
  > ...
  > 
  > Signed-off-by: dependabot[bot] &lt;support@github.com&gt;
  > Co-authored-by: dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt;



### Feature

* feat(cli): configure deployment in manifest (#9) ([`b4fd668`](https://github.com/numerous-com/numerous-sdk/commit/b4fd6680fa7bde8aa3d317279faa1b56672f24ca))

  > 
  > Add the ability to write a `deploy` section in `numerous.toml`, where `organization` and `name` (app name) can be defined, so that `numerous deploy` can be run without arguments.




## v0.4.5 (2024-06-13)

### Fix

* fix: github login (#8) ([`0d87102`](https://github.com/numerous-com/numerous-sdk/commit/0d871028fe7b64aee4e4903b3cdd9719ca05310e))

  > 
  > * separate domain and audience and authenticator constructor
  >  * use custom auth0 domain for authenticator




## v0.4.4 (2024-06-12)

### Fix

* fix(cli): prettier `push` verbose output ([`3f526ff`](https://github.com/numerous-com/numerous-sdk/commit/3f526ff619214bcf9914a0e7ed3fc3b46658c04d))




## v0.4.3 (2024-06-11)

### Fix

* fix(cli): use production auth tenant in production builds ([`7855854`](https://github.com/numerous-com/numerous-sdk/commit/7855854a7ee6a7145e88134eced09c7c81427109))




## v0.4.2 (2024-06-11)

### Fix

* fix(cli): hide verbose `deploy` output unless requested ([`f635d61`](https://github.com/numerous-com/numerous-sdk/commit/f635d616d17810f6d212ef98161fa17a9c055de9))


* fix(cli): simpler output for some access denied errors ([`2abcd15`](https://github.com/numerous-com/numerous-sdk/commit/2abcd157417b0a54a8e5ecda1ef643b1669caaa4))




## v0.4.1 (2024-06-11)

### Fix

* fix(cli): `login` and `logout` output improvements ([`6a98b6b`](https://github.com/numerous-com/numerous-sdk/commit/6a98b6bd3ca7a3f525151f17424cfc4e50530763))

  > 
  > * Display errors using functions from `output`.
  > * Add `output.PrintlnOK` for printing affirmative messages.
  > * Use non-emoji symbols with ansi coloring.




## v0.4.0 (2024-06-10)

### Feature

* feat(cli): add deploy command ([`1e8b86b`](https://github.com/numerous-com/numerous-sdk/commit/1e8b86b344c53ec671db4b0fcc0a8d423eee1892))

  > 
  > The command `numerous deploy` deploys an app to an organization, where
  > it is accessible only to authorized users.




## v0.3.6 (2024-06-04)

### Fix

* fix(cli): bug with authorization header only being added on first request ([`7a782b1`](https://github.com/numerous-com/numerous-sdk/commit/7a782b1b14f3f81fc1bb90abe8d94e84197663e4))




## v0.3.5 (2024-06-03)

### Documentation

* docs: fix workflow badges in README.md ([`d186c8a`](https://github.com/numerous-com/numerous-sdk/commit/d186c8a9cf9af94e075d6bde3b114ff86fb371d5))



### Fix

* fix(cli): remove errant quotation mark ([`449e564`](https://github.com/numerous-com/numerous-sdk/commit/449e564ddde3f66a8de8c6e2875fe39732988d2c))




## v0.3.4 (2024-05-29)

### Fix

* fix(cli): `init` prompt for using existing folder ([`db8df4e`](https://github.com/numerous-com/numerous-sdk/commit/db8df4e013698cf63dc0a513db32a202c31a8154))

  > 
  > Due to using OS unaware `path` library, absolute paths were not
  > correctly identified, causing a strange doubled path being printed on
  > Windows.




## v0.3.3 (2024-05-29)

### Fix

* fix(cli): output formatting improvements ([`3a2bc81`](https://github.com/numerous-com/numerous-sdk/commit/3a2bc81f0f310caef383e6cf1dd1e2df711d37d9))

  > * Clean paths to display better on Windows.
  > * Format paths in output with string formatting, to avoid escaping.
  > * Remove an errant quote in message.




## v0.3.2 (2024-05-28)

### Build

* build(deps): bump golang.org/x/net from 0.22.0 to 0.23.0 in /cli (#3) ([`c2472bf`](https://github.com/numerous-com/numerous-sdk/commit/c2472bf3f55000150f7e9f546607bde8ca991579))

  > 
  > Bumps [golang.org/x/net](https://github.com/golang/net) from 0.22.0 to 0.23.0.
  > - [Commits](https://github.com/golang/net/compare/v0.22.0...v0.23.0)
  > 
  > ---
  > updated-dependencies:
  > - dependency-name: golang.org/x/net
  >   dependency-type: indirect
  > ...
  > 
  > Signed-off-by: dependabot[bot] &lt;support@github.com&gt;
  > Co-authored-by: dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt;


* build(deps): bump github.com/lestrrat-go/jwx in /cli (#1) ([`6b9c01a`](https://github.com/numerous-com/numerous-sdk/commit/6b9c01a21fc857c66e13143cd9def4b685dc643e))

  > 
  > Bumps [github.com/lestrrat-go/jwx](https://github.com/lestrrat-go/jwx) from 1.2.28 to 1.2.29.
  > - [Release notes](https://github.com/lestrrat-go/jwx/releases)
  > - [Changelog](https://github.com/lestrrat-go/jwx/blob/v1.2.29/Changes)
  > - [Commits](https://github.com/lestrrat-go/jwx/compare/v1.2.28...v1.2.29)
  > 
  > ---
  > updated-dependencies:
  > - dependency-name: github.com/lestrrat-go/jwx
  >   dependency-type: direct:production
  > ...
  > 
  > Signed-off-by: dependabot[bot] &lt;support@github.com&gt;
  > Co-authored-by: dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt;



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

* feat(cli): `push` reads `.env` and sends secrets ([`c019861`](https://github.com/numerous-com/numerous-sdk/commit/c019861e47d99b2109e0091cd57f29e0c950fd72))

  > 
  > Read and parse `.env` in the app directory, and send the parsed
  > environment to the server for it to configure the resulting app
  > deployment with the secrets from the `.env`.




## v0.2.0 (2024-05-07)

### Feature

* feat(cli): create `.app_id.txt` to store App ID ([`7a8f682`](https://github.com/numerous-com/numerous-sdk/commit/7a8f68238319e714c94851b43904f0e5a3e8f703))

  > 
  > Includes backwards compatibility for projects initialized with `.tool_id.txt`. Also includes various improvements to output
  > formatting related to files bootstrapping.



### Fix

* fix(cli): improved output for commands, and minor refactors ([`15bdc92`](https://github.com/numerous-com/numerous-sdk/commit/15bdc92e8747cf40e4a1ebe3139c4a317ae8b784))

  > 
  > Use common error printing functions, and improve the phrasing of some error and informative messages.


* fix(cli): bug `numerous push` fails due to only reading deprecated App ID file ([`f80458b`](https://github.com/numerous-com/numerous-sdk/commit/f80458b3a49aa8e2e1823d82c3e8a02274214695))

  > 
  > Use the common function read the App ID, which can read from `.app_id.txt`, since it is now created by the `numerous init`command




## v0.1.3 (2024-05-06)

### Fix

* fix(cli): login and logout output ([`becb0ae`](https://github.com/numerous-com/numerous-sdk/commit/becb0ae7145f4a46a3b682be961f70ef681029cf))




## v0.1.2 (2024-05-06)

### Fix

* fix(cli): improve &#39;numerous log&#39; output ([`1b3e7e5`](https://github.com/numerous-com/numerous-sdk/commit/1b3e7e52545d290b330eba6690a0191fabc1df14))

  > 
  > * A bit of refactoring
  > * Adds flag to print timestamps
  > * Adds standardized error printing functions




## v0.1.1 (2024-05-06)

### Documentation

* docs: references to examples in README.md ([`e12d309`](https://github.com/numerous-com/numerous-sdk/commit/e12d30997ce6683708dfa1af326d625cc9583119))



### Fix

* fix(cli): improve &#39;init&#39; output, and use &#39;.app_id.txt&#39; ([`75798d1`](https://github.com/numerous-com/numerous-sdk/commit/75798d1827eeba9dc797057e05db867523d78af7))

  > 
  > * Make output of the &#39;numerous init&#39; command more readable, friendly and colorful.
  > * Read &#39;.app_id.txt&#39; for the App ID, falling back to the old &#39;tool_id.txt&#39;.
  > * Added a test helper for writing to a file.
  > * Removed a println in the bootstrap code.
  > * Fixed some print statements that were lower cased.




## v0.1.0 (2024-05-03)

