# CHANGELOG



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
