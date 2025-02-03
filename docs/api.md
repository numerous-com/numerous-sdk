# Enabling/Setting up API Access

Numerous allows you to expose your apps functionality not just through a web UI, but also as an API. API capability is useful when your users want to include your app’s functionality in their own scripts and applications.

If you would like to enable your users to have access to your app via API requests, follow these steps:

1\. Manually update the TOML file and add the following lines:

```
# numerous.toml

# existing manifest content ...

[api]
  enabled = true
```

If you would like to include API docs for your app, you can include a docs path similar to this (where `/docs` needs to be changed to the actual docs route):

```
# numerous.toml

# existing manifest content ...

[api]
  enabled = true
  docs_path = "/docs"

```

2\. Deploy the app using the `numerous deploy` command
<br/>3\. The backend will update the app "state" and will enable the API access.
<br/>4\. Once you have enabled API access, users can find the API docs and option to create API keys on the app card. Find more information about API Keys on the [platform documentation page](https://www.numerous.com/docs).
