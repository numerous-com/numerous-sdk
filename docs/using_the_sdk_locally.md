# Using the SDK locally

When you deploy your app to the Numerous Platform, the SDK will automatically
connect to the server, so that collections, documents and files are stored
remotely.

When run locally, the SDK uses "local mode", but you can also connect the SDK
to the platform, which can be useful if you want to create scripts that update
or generate data for your app.


## Using local mode

By default when you use the SDK locally on your local machine, it will be in
"local mode", and store collections, documents and files on the file system.
This is primarily intended for testing.

By default the folder `collections` is used to store the data, but if you set
the environment variable `NUMEROUS_COLLECTIONS_BASE_PATH`, the value will be
used as the folder for storing the data.


## Connecting the SDK to the platform locally

In order to connect the SDK to the platform locally it needs to be configured
with an access token, and an organization ID. You can generate these values
using the CLI!


### Generating an access token

First create an access token. Here we create an access token named `sdk-local`.

```{ .optional-language-as-class .no-copy }
$ numerous token create -n sdk-local
✓ Created personal access token "sdk-local": num_bS6KjPnMA52R3O73xHhWoekbUUXwMI1fVrp40LNe
Make sure to copy your access token now. You won't be able to see it again!
```


### Finding the organization ID

Now we find the organization ID by listing all organizations we are member of
using the CLI.

```{ .optional-language-as-class .no-copy }
$ numerous organization list
Name: John Doe's Organization
Slug: john-doe-abcd1234
Role: ADMIN
ID:   f656cb0bba274a3193b121d0c3415d1b
```

### Creating a `.env` file to store credentials

We create a `.env` file which contains the values we just found.

Of course you should substitute these example values with your actual access
token and organization ID.

```
NUMEROUS_API_ACCESS_TOKEN=num_bS6KjPnMA52R3O73xHhWoekbUUXwMI1fVrp40LNe
NUMEROUS_ORGANIZATION_ID=f656cb0bba274a3193b121d0c3415d1b
```

!!! warning
    Remember to make sure that the `.env` file with your credentials is not
    checked into your version control.
    
    `numerous init` will have added a `.gitignore` rule that ignores `.env`
    already, and it is also added to the `exclude` section of your app manifest
    `numerous.toml` by default.

### Loading the credentials in your python script

First we need to install `python-dotenv`. Do this with `pip` in the terminal, or
perhaps with the integrated tooling of your IDE.

Now we update the code to load the `.env` file. Below we use the `python-dotenv`
package to load the `.env` file we just created. In this case we only want to
load the `.env` file locally, so we wrap the loading logic in a `try-except`
block, so that it will still work when run remotely.


```python
from numerous.collections import collection

# 
try:
    import dotenv
    dotenv.load_dotenv()
except ModuleNotFoundError:
    pass

col_ref = collection("my-collection)

# use the collection reference ...
```