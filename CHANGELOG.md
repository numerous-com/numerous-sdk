# CHANGELOG


## v0.24.0 (2024-10-03)

### Feature

* feat: reduce package size with platform specific distributions (#38) ([`057f124`](https://github.com/numerous-com/numerous-sdk/commit/057f1241726631e5bdc5ed9b83f57ecfeab1884d))

  > 
  > Reduces package size by a factor of 6: ~66MB to ~11MB.
  > 
  > Changes how the CLI binaries are added to the python package, and how
  > the python executable launches the binary.
  > 
  > 1. The binaries are now located in `numerous/cli/bin`.
  > 2. The python script tries to launch `numerous/cli/bin/cli` if it
  >    exists. It exists if the package is installed from a platform
  >    specific wheel.
  > 3. Otherwise the python script will attempt to match the the current OS
  >    with a binary, similar to the approach until now. This is meant for
  >    non-platform specific wheels.
  > 
  > The build process is changed. Wheels and source distribution are created
  > by `scripts/build_dists.sh`, which can create a bdist for each relevant
  > platform that we build CLI binaries for. The CLI binary directory
  > `numerous/cli/bin` is cleared, and the relevant CLI binary is copied
  > into `numerous/cli/bin/cli`, and the corresponding wheel is created.
  > 
  > For the platform non-specific wheel and the source distribution all
  > binaries are added, as before.




## v0.23.1 (2024-10-02)

### Fix

* fix(cli): handle encoded requirements.txt files (#40) ([`6ab531f`](https://github.com/numerous-com/numerous-sdk/commit/6ab531f638f6c54602ec7158db3cdcbaa1f4c2aa))

  > 
  > * Handle `requirements.txt` files encoded with `utf-8`, `utf-16` and `utf-32` regardless
  >   of endianness, by detecting the corresponding BOMs.
  > * Read `requirements.txt` files with both `lf` and `crlf` line termination.
  > * Preserve `crlf` line termination by checking if any lines are terminated by
  >   `crlf`, and then using `crlf` when re-encoding if any `crlf` line termination was found.
  > 
  > Fixes #37




## v0.23.0 (2024-10-02)

### Feature

* feat(cli): version check (#39) ([`79384f6`](https://github.com/numerous-com/numerous-sdk/commit/79384f6c99a9785499709b95f878e04af3ebd115))

  > 
  > Adds CLI compatibility check so depending on the version it:
  > * continues to work;
  > * shows a warning and continues to work;
  > * stops with an error message.




## v0.22.0 (2024-10-01)

### Feature

* feat(python-sdk): local file system collections (#36) ([`8e18f76`](https://github.com/numerous-com/numerous-sdk/commit/8e18f762ecd4ad8e124f1fe4cb4c01196cf91581))

  > 
  > Enables using the collections features locally, by using an implementation that stores
  > collections and documents in directories on the local file system, when running locally.
  > 
  > Set the environment variable `NUMEROUS_COLLECTIONS_BASE_PATH` to define
  > where to put collections and documents. Defaults to the directory `collections` in
  > the working directory.
  > 
  > ---------
  > 
  > Co-authored-by: lasse-numerous &lt;lasse.thomsen@numerous.com&gt;




## v0.21.0 (2024-10-01)

### Feature

* feat(cli): command `app list` for listing apps (#34) ([`1f93d82`](https://github.com/numerous-com/numerous-sdk/commit/1f93d82386f9f0d9831b47e52ab5f52ee59a0b79))




## v0.20.0 (2024-09-30)

### Feature

* feat(cli): command `token list` to list personal access tokens (#35) ([`305e989`](https://github.com/numerous-com/numerous-sdk/commit/305e9899e038a91253638d601db59a7ce79de337))




## v0.19.2 (2024-09-30)

### Fix

* fix(python-sdk, cli): properly return error exit codes ([`dd61ff7`](https://github.com/numerous-com/numerous-sdk/commit/dd61ff76fdd7bc55cc8871d53e48eb378708a7f1))




## v0.19.1 (2024-09-25)

### Fix

* fix(python-sdk): organization environment variable and default client ([`29854ad`](https://github.com/numerous-com/numerous-sdk/commit/29854adf99555e94fc9ef7050e8ffb290ff49138))

  > 
  > * Use the correct environment variable `NUMEROUS_ORGANIZATION_ID`.
  > * Define exception types for relevant exceptions in `_client` module.
  > * Raise exception if organization ID is not configured.
  > * Default to singleton client.




## v0.19.0 (2024-09-25)

### Feature

* feat(python-sdk): nested collections and documents (#32) ([`a6efc63`](https://github.com/numerous-com/numerous-sdk/commit/a6efc63a5fc5615759cd3a058671b471e076590b))

  > 
  > Documents are indexed by a key in a collection, and contain JSON data:
  > * Create documents in a collection
  > * Manage tags for a document
  > * Iterate over documents in a collection (with filtering by tags possible)
  > * Read document data
  > 
  > Additionally nested collections are introduced.




## v0.18.0 (2024-09-25)

### Feature

* feat(cli): command `token revoke` to revoke personal access tokens (#33) ([`0272afb`](https://github.com/numerous-com/numerous-sdk/commit/0272afb8c39a9c7b23c5c265db515107e956f327))




## v0.17.0 (2024-09-23)

### Feature

* feat(cli): add `version` command to print version number (#30) ([`c840721`](https://github.com/numerous-com/numerous-sdk/commit/c8407218b2f249a4a60032af908c0f146b8d6495))




## v0.16.1 (2024-09-20)

### Fix

* fix(cli): dockerfile app example uses provided port number ([`57f910b`](https://github.com/numerous-com/numerous-sdk/commit/57f910b39153b52b7e0f9bc57e183eeb5b5edb8f))

  > 
  > Fixes a bug where the initialized docker file example used a hardcoded
  > port number instead of the port number provided in the wizard or by
  > command line arguments.




## v0.16.0 (2024-09-20)

### Feature

* feat(cli): initialize apps based on docker builds (#31) ([`b72af15`](https://github.com/numerous-com/numerous-sdk/commit/b72af150d0a606e6d1562ba138d883b141e78673))

  > 
  > * Enables creating apps based on Dockerfiles.
  >   - Adds new option in wizard.
  >   - Adds new command line flags.
  > * Introduces new `numerous.toml` format.
  > * Backwards compatible with apps initialized previously.




## v0.15.3 (2024-09-19)

### Fix

* fix(cli): `download` command handles access denied error ([`0418a7e`](https://github.com/numerous-com/numerous-sdk/commit/0418a7ea93a2fb2a714fdb7989beca7550c0e360))




## v0.15.2 (2024-09-19)

### Fix

* fix(cli): properly print errors for `delete` and `download` commands ([`7c9a0d0`](https://github.com/numerous-com/numerous-sdk/commit/7c9a0d08db7a0897bcb4e0720206e5f6d2f2e52e))




## v0.15.1 (2024-09-16)

### Fix

* fix(cli): streamline error output (#28) ([`6094e00`](https://github.com/numerous-com/numerous-sdk/commit/6094e004ca29e1f02a19203a4d0a1d95fa7e15a6))

  > 
  > Improve consistency in how errors are presented to users.
  >  
  > Do not re-print errors in the command wrapper that have already been printed internally, but keep the errors to signal that the exit code should be `1`.
  > 
  > Only show usage for command argument errors, not errors happening internally in the command functions.




## v0.15.0 (2024-09-04)

### Feature

* feat(python-sdk): access and create collections (#22) ([`be13de3`](https://github.com/numerous-com/numerous-sdk/commit/be13de3fbd925cd40780218139b27dfd74c49d53))

  > 
  > Adds `numerous.collection` function which creates or reads the collection,
  > and allows recursive collection access.




## v0.14.3 (2024-08-30)

### Fix

* fix(cli): `token create` displays an error if token name is not supplied ([`dead63f`](https://github.com/numerous-com/numerous-sdk/commit/dead63f7c5f83c2d41dd5539e7b2337abdaf30a9))




## v0.14.2 (2024-08-30)

### Fix

* fix(cli): allow running commands requiring authentication with access token ([`61182cd`](https://github.com/numerous-com/numerous-sdk/commit/61182cdbeb13b717f46d29b5dbf7a770dd49137b))




## v0.14.1 (2024-08-29)

### Fix

* fix(cli): force a rebuild and release of `numerous token` feature ([`1198ca8`](https://github.com/numerous-com/numerous-sdk/commit/1198ca88b8b6ccf6619cf5e33c01215d1eda64ff))




## v0.14.0 (2024-08-29)

### Feature

* feat(cli): command `token create` to create personal access tokens (#27) ([`b0fd9fe`](https://github.com/numerous-com/numerous-sdk/commit/b0fd9feb91062687461e947ef4a3c3f009c443ed))

  > 
  > Adds a command `numerous token create` which can be used to create a personal
  > access token. The personal access token can be used to automate the CLI by setting
  > the environment variable `NUMEROUS_ACCESS_TOKEN`.




## v0.13.0 (2024-08-21)

### Feature

* feat(cli): add option to initialize a Panel app (#25) ([`fab32cc`](https://github.com/numerous-com/numerous-sdk/commit/fab32cc9a04d11a2b3925ae66de34c55efb0770e))




## v0.12.2 (2024-08-21)

### Fix

* fix(python-sdk): consistent isort linter config ([`3578ef0`](https://github.com/numerous-com/numerous-sdk/commit/3578ef00d444e58d7369890631ed71eba5e97022))




## v0.12.1 (2024-08-09)

### Fix

* fix(cli): typos in `login` and `download` command descriptions ([`7752c16`](https://github.com/numerous-com/numerous-sdk/commit/7752c1643d78e4e281d8d8de5410a53000f6f375))




## v0.12.0 (2024-08-09)

### Feature

* feat(cli): `download` command for downloading app sources (#24) ([`584ce22`](https://github.com/numerous-com/numerous-sdk/commit/584ce22c90bfc56e8c9f62eeacfd033216b368d4))




## v0.11.2 (2024-08-07)

### Fix

* fix(cli): properly check if cli is logged in ([`2ece1f5`](https://github.com/numerous-com/numerous-sdk/commit/2ece1f5d8881d28b053ddeed5d51fa67b5c10113))

  > 
  > Fixes a condition that was reversed, causing the login command to only
  > allow logging in when the CLI is already logged in.




## v0.11.1 (2024-08-06)

### Fix

* fix(cli): `init` command error printing improvements (#23) ([`c4c2adc`](https://github.com/numerous-com/numerous-sdk/commit/c4c2adc41de9afb214642295225133279eb83643))

  > 
  > * Do not print initialization preparation errors twice for the non-legacy command.
  > * Do not print interrupts as errors.
  > * Improve test readability




## v0.11.0 (2024-08-06)

### Feature

* feat(cli): feedback message before commands, remove `report` command (#20) ([`4b8e612`](https://github.com/numerous-com/numerous-sdk/commit/4b8e612d33764f08db7404cebd122743ddc7e29e))

  > 
  > * Ask users for feedback before every command with a 10% probability
  > * Removes the now redundant `report` command in favor of simply printing the message
  >   when running other commands.




## v0.10.3 (2024-07-10)

### Fix

* fix: app error message ([`a8c169e`](https://github.com/numerous-com/numerous-sdk/commit/a8c169e077bbaf32672e5ed30308891337531174))



### Unknown

* Update README.md (#18) ([`82fc974`](https://github.com/numerous-com/numerous-sdk/commit/82fc974f0b66ab27c5287c96abefb28372ba33a4))




## v0.10.2 (2024-07-02)

### Fix

* fix(cli): rename `deploy` app slug flag ([`ce0a24f`](https://github.com/numerous-com/numerous-sdk/commit/ce0a24f5a80fdcb9381655fee2391f56763c4d2e))




## v0.10.1 (2024-07-02)

### Fix

* fix(cli): improve error messages (#19) ([`4bfeff3`](https://github.com/numerous-com/numerous-sdk/commit/4bfeff356e4b349f8b778991aeee521eaa01fcfe))

  > 
  > * Do not print `delete` errors twice
  > * Print specific error messages for common cases
  > * Refactor error printing and move printing to `cmd` package




## v0.10.0 (2024-06-28)

### Feature

* feat(cli): promote organization app commands, and add `legacy`  namespace (#17) ([`86f951d`](https://github.com/numerous-com/numerous-sdk/commit/86f951d346434b34cce850f42d53922c9fd164a8))

  > 
  > * Promote commands related to organization apps (`deploy`, `delete`, `logs`) to be root commands.
  > * Rename &#34;app name&#34; concept of new app commands to &#34;app slug&#34; to avoid confusion.
  > * Move existing commands for managing apps with `.app_id.txt` to the `legacy` namespace.
  > * Add new `numerous init` command  that does not register a legacy app in the Numerous server,
  >   or create `.app_id.txt`. Move existing `init` to the `legacy` namespace.
  > * Add helpful messages to direct the user to the correct commands.
  > * For `numerous deploy`, `numerous logs`, and `numerous delete` if no app slug is provided
  >   through CLI arguments, or in the manifest, use a sanitized version of the app display name
  >   from the manifest




## v0.9.4 (2024-06-27)

### Fix

* fix(cli): change report command and add tests (#6) ([`9d7198b`](https://github.com/numerous-com/numerous-sdk/commit/9d7198b4844e32e34a329942c72d49866aeb246c))

  > 
  > test(cli): add report command tests
  > 
  > test(cli): improves test logic and clarification
  > 
  > test(cli): changes error messages to use the numerous standard output
  > 
  > fix(cli): removes bool reference from wsl check
  > 
  > test(cli): changes after code review
  > 
  > refactor: changes wsl identification function




## v0.9.3 (2024-06-26)

### Fix

* fix(cli): segfault in  without a  section in manifest ([`0478046`](https://github.com/numerous-com/numerous-sdk/commit/04780461ba85fbad76ba4801978793aaaad3e740))




## v0.9.2 (2024-06-26)

### Fix

* fix(cli): `deploy` should fall back to manifest deployment section ([`28044e1`](https://github.com/numerous-com/numerous-sdk/commit/28044e1513a1b248bd651000215d7a37974213d1))

  > 
  > Fixes a bug where the manifest `deploy` section was not used in lieu of arguments.


* fix(cli): better help message for access denied error in `deploy` ([`7c02663`](https://github.com/numerous-com/numerous-sdk/commit/7c02663ce759897c27a794aaea26f892881e0967))


* fix(cli): build with production configuration ([`46de7fb`](https://github.com/numerous-com/numerous-sdk/commit/46de7fb8e4e435df8de2656e9e84ac7c327ddf8f))

  > 
  > After refactoring the module and package structure, the linking of
  > configuration variables was broken.




## v0.9.1 (2024-06-25)

### Documentation

* docs: remove references to `cli` directory ([`259aa9f`](https://github.com/numerous-com/numerous-sdk/commit/259aa9f0c38993a2e5f981761dc4d389a451dab6))



### Fix

* fix(cli): help message for app logs (#16) ([`a9b5a48`](https://github.com/numerous-com/numerous-sdk/commit/a9b5a486f3333e7fb4650b7247e6df140347cfd5))




## v0.9.0 (2024-06-24)

### Feature

* feat(cli): add `--version` and `--message` flags to `app deploy` (#15) ([`d41290b`](https://github.com/numerous-com/numerous-sdk/commit/d41290bd49b74c3da20454da79ef380f9ed07882))

  > 
  > Allow marking deployed versions with a version number and a message.




## v0.8.0 (2024-06-19)

### Feature

* feat(cli): add `app delete` command (#13) ([`e1c4536`](https://github.com/numerous-com/numerous-sdk/commit/e1c4536774bcb73cbdbbccfc6008e0b017c87625))

  > 
  > Adds the command `numerous app delete` which deletes an app from Numerous, all associated information, and stops all related workloads.




## v0.7.0 (2024-06-18)

### Feature

* feat(cli): add `app logs` command (#12) ([`6ab6c24`](https://github.com/numerous-com/numerous-sdk/commit/6ab6c2426c606ace3474a3b71469ad1f20141d82))

  > 
  > Adds the command numerous app logs which prints the app logs from a deployment of an app.




## v0.6.2 (2024-06-17)

### Fix

* fix(cli): `app deploy` optional organization/name flags ([`f061947`](https://github.com/numerous-com/numerous-sdk/commit/f0619470605bec7b06facbb65a07749cbee0f7b0))

  > 
  > Makes it possible to load deployment configuration from the manifest
  > by making it possible to call the command without `--organization` and `--name` flags.




## v0.6.1 (2024-06-17)

### Fix

* fix(cli): `app deploy` help possessive apostrophe, and slug example ([`35fe75d`](https://github.com/numerous-com/numerous-sdk/commit/35fe75da7499dbc1d8e525bd5ab620f6ae349697))




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

