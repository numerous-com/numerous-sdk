# Command-Line Interface (CLI) Reference

The Numerous CLI Reference provides information about currently available
commands and features.

## Install

First, install the `numerous` Python package, which includes the CLI.

!!! note
Before you install the Numerous CLI, we recommend that you create and activate
a [Python virtual environment](https://docs.python.org/3/tutorial/venv.html#creating-virtual-environments).

In your virtual environment (or without one, if you prefer), run the following
command:

```
pip install numerous
```

The `pip install numerous` command installs the Numerous CLI to your computer.
To see if the installation was successful, run:

```
numerous
```

Available commands and a help message will be displayed if the installation
process was successful. <br/>
If you encounter an error message, check for errors and try again. If the
problem persists, please reach out to us at
[support@numerous.com](mailto:support@numerous.com).

## Initialize

```
numerous init
```

The initialization command `numerous init` bootstraps an app in the current
directory. Specifically, it creates a file, `numerous.toml`, which contains the
app configuration.

When you run the command, a wizard will help you fill in your app's name, a
description, and the app engine you are using.

It also specifies the app entrypoint (your Python file), and your
`requirements.txt`. Our server uses this information later when your app is
being built.

.. note::
It is important that you are in the right directory when you call `numerous
  init`, since it initializes that directory for an app.

Alternatively if you want to initialize your app in the directory
`my-projects/my-app` you can run the command with the directory as an argument
`numerous init my-projects/my-app`.

You can also specify the information from the wizard directly through flags to
the command, for example: `numerous init --name "My Project"`.

To see all available flags:

```
numerous init --help
```

## Log in / Sign up

```
numerous login
```

The `numerous login` command allows you to log in or sign up to the CLI.

Once the command is executed, you will be prompted to complete authentication
with an 8-letter code (for example: XXXX-XXXX). You can press `enter` to open
the browser or `control + c` to quit.

Once the authentication code is verified, you will be prompted to either log in
or sign up.

Once completed, the text in the terminal will confirm that you are successfully
logged in.

## Log out

```
numerous logout
```

The `numerous logout` command allows you to log out of the CLI.
Once the command is executed, text in the terminal will confirm that you are
successfully logged out.

## Update

```
pip install --upgrade numerous
```

When using the Numerous platform, you will be informed when there is an update
available. You will then be prompted to run the `upgrade` command.

## Create an organization

```
numerous organization create
```

The `numerous organization create` command creates a new organization on the Numerous platform.

When executing the command, you will be taken through a setup wizard where you can create a name
for your organization. If you would like to exit the wizard at any point, press ctrl + c.
After completing the wizard, you'll get a URL that links to the organization page on the Numerous web platform.

Visit [the platform docs](www.numerous.com/docs) to read more about organizations.

!!! note
You must be logged in to the CLI to use `numerous organization create`.

## List organizations

```
numerous organization list
```

The `numerous organization list` command lists all the organizations that you are an administrator or member of.
It can be useful for finding the organization slug of an organization to deploy an app to.

Visit [the platform docs](www.numerous.com/docs) to read more about organizations.

!!! note
You must be logged in to the CLI to use `numerous organization list`.

## Deploy

```
numerous deploy
```

The `numerous deploy` command is used to upload your app code and launch it on
the Numerous platform. Whenever you deploy, the “app slug” and "organization
slug" specify _where_ the app should be deployed to.

After you have deployed your app, it can be accessed on your organization's app
gallery.

If you want to deploy your app using the “app slug” `my-app` to the organization
with the slug `my-organization-abcd1234`, you can use the command line flags below:

```
numerous deploy --app my-app --organization my-organization-abcd1234
numerous deploy -a my-app -o my-organization-abcd1234
```

Or, you can add a `[deploy]` section to the end of `numerous.toml`, after which
you do not need to specify the command line arguments.

```toml
# existing manifest content ...

[deploy]
app="my-app"
organization="my-organization-slug-abcd1234"
```

## Download

```
numerous download
```

The `numerous download` command downloads an app from the Numerous
platform to your local drive. This may be useful when utilizing a
template app that you would like to edit or if you want to access the
code currently running in an app to debug an issue. Use `numerous download`
followed by the organization slug (`-o`) and app slug(`-a`) arguments.

<br /> For example:

```
numerous download -o my-org-slug -a my-app-slug
```

Once it is downloaded, you can edit using your local drive.
By default, it will download the source code into the folder of the same name as the given app slug in this example `my-app-slug`.
You can specify a folder argument as the last argument to select a different folder.

```
numerous download -o numerous-abcd1234 -a my-app-slug
# download to current folder numerous download -o numerous-abcd1234 -a my-app-slug another-app-folder
# download to "another-app-folder"
```

## Delete

```
numerous delete
```

The `numerous delete` command deletes an app that you no longer want to have on
the Numerous platform. The app and its associated resources will be deleted once
the command has been executed.

This command only deletes the selected app from the Numerous servers and does
not affect your local folder.

Like the `numerous deploy` command, you must specify an “app slug” and an
“organization slug” to specify which app you want to delete, or you can
[configure the deployment configuration](#default-deployment-configuration).

```
numerous delete --app my-app --organization my-organization-abcd1234
numerous delete -a my-app -o my-organization-abcd1234
```

## Logs

```
numerous logs
```

The `numerous logs` command lists logs written by the deployed app to the
standard output and standard error file descriptors. It is useful for debugging
if your app prints error messages, or other information.

Logs from the last hour are printed, and then the output is “followed”, meaning
the command waits for new output from your app until you cancel the command with
`control + c`.

```
numerous logs --app my-app --organization my-organization-abcd1234
numerous logs -a my-app -o my-organization-abcd1234
```

Like the `numerous deploy` command, you must specify an “app slug” and an
“organization slug” to specify which app you want to delete, or you
can [configure the default deployment configuration](#default-deployment-configuration).

## Personal access tokens

With personal access tokens you can use the CLI in scripts, such as for
continuous delivery of your app, e.g. [with GitHub actions](/github-actions).

### Creating a personal access token

Personal access tokens are created with `numerous token create`. You must
specify a name and have the option to add a description.

The name must be no longer than 40 characters.

```
numerous token create --name my-access-token --description "A description of my access token"
```

### Using a personal access token

In order to use a personal access token, it needs to be assigned to the
environment variable `NUMEROUS_ACCESS_TOKEN`. Below the `numerous deploy`
command is called with a personal access token for authentication.

```bash
NUMEROUS_ACCESS_TOKEN=num_tKYaoAVoGUaJ1s5UG6QK8vwnNG1n03g8allA7zSL numerous deploy
```

## Configure

```toml
name="Energy System Simulator"
description="An app for simulating energy systems."
port=8080
cover_image="my_cover.png"
exclude=["*venv", "venv*", ".git"]

[python]
version="3.11"
library="streamlit"
app_file="app.py"
requirements_file="requirements.txt"
```

A configuration file called `numerous.toml` is automatically created when the
`numerous init` command is executed. Certain app elements can be modified by
editing fields in your configuration file, such as renaming your app, changing
your library, or excluding additional files.

For example, in the example configuration above, we could change the library
from Streamlit to Marimo by updating the library field from “streamlit” to
“marimo”.

!!! note
We recommend that you check `numerous.toml` into your version control system.

#### Exclude certain files and folders

There is a size limit og 5 gigabytes when uploading app sources to Numerous.
We recommend excluding large files or folders when uploading into your project
directory.

In `numerous.toml` the `exclude` field contains a list of rules, that matches
files in your app folder. Any file matched by a rule is excluded from the
resulting app archive.

Certain files are excluded by default, specifically any folders
matching `*venv` and `venv*` to filter out virtual environments as well as
`.git`. This is done to avoid uploading the entire version control history with
every push.

You can update the `exclude` field to exclude more files. For example, if you
have a folder `testdata` with a lot of large `.csv` files, you might add an
exclude rule for `testdata/*.csv`, which will exclude any `.csv` file in the
`testdata` folder. This rule is appended to the default rules in the example
below:

```toml
# numerous.toml
exclude = ["*venv", "venv*", ".git", "testdata/*.csv"]
```

#### Building your app from a Dockerfile

You can exchange the `[python]` section in `numerous.toml` with a `[docker]`
section to define a build with a `Dockerfile`.

```toml
name="Dockerfile App Example"
description="An example for showing how to build an app from a Dockerfile."
port=8080
cover_image="my_cover.png"
exclude=["*venv", "venv*", ".git"]

[docker]
dockerfile="Dockerfile"
context="."
```

See the [documentation of Dockerfile](https://www.numerous.com/docs/app-engines/dockerfile) for more
information.

#### Default deployment configuration

In the manifest file `numerous.toml`, it is possible to specify a default
deployment target. Add a `[deploy]` section, populate the `organization`
field with the organization slug identifier, and the `app` field with an app
slug identifier.

By default, the app slug identifier is created from the app display name (defined
under the `name` converted into a slug). It is lower-cased, all
spaces are replaced with dashes, and any special characters are stripped.

```toml
# previous configuration

[deploy]
app="my-app"
organization="my-organizations-slug"
```

## Legacy commands

The first versions of Numerous CLI identified the app you are working on with an
`.app_id.txt` file in the app directory, which contains a unique ID for your
app, which acts as the key.

This way of identifying apps is being discontinued and apps must now
be deployed to an organization.

Below is the original documentation for the legacy commands, which can still be
used for the time being. However, we recommend new users instead use the commands listed above.

### Initialize (legacy)

```
numerous legacy init
```

The initialization command `numerous legacy init` bootstraps an app in the current
directory. Specifically, it creates a file, `numerous.toml`, which contains the
app configuration.

When you run the command, a wizard will help you fill in your app's name, a
description, and the app engine you are using.

It also specifies the app entrypoint (your Python file), and your
`requirements.txt`. Our server uses this information later when your app is
being built.

.. note::
It is important that you are in the right directory when you call `numerous
  init`, since it initializes that directory for an app.

You can also specify the information from the wizard directly through flags to
the command, for example: `numerous legacy init --name "My Project"`.

To see all available flags:

```
numerous legacy init --help
```

### Push (legacy)

```
numerous legacy push
```

The `numerous legacy push` command uploads your app directory to our server. It
will then be built, made available through a shared URL, and be printed to the
terminal.

You can use this URL to open your app in the browser and send it
to those you want to share your application with.

By default, `numerous legacy push` uploads the app from the current folder, but
you can also specify the path to the app that you want to push. For example, if
the app is in the folder `apps/myapp`, you would run
`numerous legacy push apps/myapp`.

!!! note
We currently support "shared" apps, which means that the app instance is
shared between all users. This means that any global state in the application
is also shared between everyone who uses the application.

### List apps (legacy)

```
numerous legacy list
```

The `numerous legacy list` command provides an overview of your available apps.
Once the command is executed, you will be presented with a table that includes
all of your available apps. The table includes each app's description, time
created, its shareable URL, and public app gallery status.

Both the `numerous legacy list` and `numerous legacy push` commands output a
shareable URL for your app. Using `numerous legacy list` to obtain the shareable
URL for your app may be helpful to use if you are not yet ready to push your
app.

!!! note
You must be logged in to the CLI to use `numerous legacy list`.

### Publish (legacy)

```bash
numerous legacy publish
```

The `numerous legacy publish` command makes your app available on the Public App
Gallery at https://numerous.com/apps. Apps that have been published to the Public
App Gallery are freely available for use.

Once the command is executed, a public link to the app will be generated:

```text
https://numerous.com/app/public/<some_hash>
```

This link is different from the link generated by the `numerous legacy push`
command.

### Unpublish (legacy)

```
numerous legacy unpublish
```

The `numerous legacy unpublish` command removes all access to an app through the
public link generated by the `numerous legacy publish` command. Additionally,
the app will no longer be featured on the Public App Gallery at
https://numerous.com/apps.

Unpublishing an app will not affect access through the link generated by the
`numerous legacy push` command.

### Delete (legacy)

```
numerous legacy delete
```

The `numerous legacy delete` command deletes an app that you no longer want to
have on the Numerous platform. The app and its associated resources will be
deleted once the command has been executed.

This command only deletes the selected app from the Numerous servers and does
not affect your local folder.

!!! note
This command can only be executed if `numerous legacy init` was previously executed.
