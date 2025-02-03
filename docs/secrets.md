# Using Secret Variables

Numerous supports “secret variables” that make it possible to store your credentials securely.
Secret variables refer to passwords, credentials, or other information you would not like to be shown in the source code.

## Add your secret values to `.env`

In order to define the secret variables used in your app, you must create a file
called `.env` and add it to your app folder. This is also the folder where the
manifest file `numerous.toml` is stored.

For example, it might look like the following:

```txt filename=".env"
MY_SECRET=A secret value
```

.. note::It is important to ensure that `.env` is ignored in your version control (e.g.
git). We also recommend that you exclude it from any uploads to the Numerous
platform.

Apps initialized with Numerous CLI version greater than 0.3.1 have this setup
by default, but existing apps should update their `.gitignore` and
`numerous.toml`. See instructions for updating `numerous.toml` [here](/cli#configure).

## Access secrets in environment variables

In the app code, secrets are accessible as environment variables. In Python, you
can access environment variables using the
[`getenv`](https://docs.python.org/3/library/os.html#os.getenv) function in the
[`os`](https://docs.python.org/3/library/os.html) module.

As an example, consider this simple Streamlit app that accesses the previously
defined `MY_SECRET`.

```python filename="app.py"
import os
import streamlit as st

st.title("A test of secrets")

st.text(f"My secret: {os.getenv('MY_SECRET')}")
```

The above code will result in this app being rendered:
![](/docs/static/secrets-app-screenshot.png)

## Deploy your app

In order to re-deploy your app with the specified secrets, deploy your app
after updating `.env`.

<Callout>
  If you remove any secret variables from `.env`, they will no longer be
  available to your app.
</Callout>
