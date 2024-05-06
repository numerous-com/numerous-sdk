# CHANGELOG



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
