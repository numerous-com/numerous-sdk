# WSL Support for Numerous CLI

This directory includes support for running the Numerous CLI in Windows Subsystem for Linux (WSL) environments where the system keyring may not be available or functional.

## Automatic Fallback (Recommended)

**NEW**: The Numerous CLI now automatically detects when keyring operations fail and falls back to file-based credential storage. This means:

- **No manual configuration needed** - The CLI will automatically switch to file-based storage when it detects keyring issues
- **Seamless experience** - You'll get a notification when fallback occurs, but authentication will continue to work
- **Cross-platform compatibility** - Works on WSL, Linux without GUI, Docker containers, and other environments where keyring access is limited

When automatic fallback is triggered, you'll see a message like:
```
Keyring access failed, falling back to file-based credential storage
Using file-based credentials storage: /home/username/.numerous/credentials.json
Note: On WSL/Linux, you can avoid this by setting NUMEROUS_LOGIN_USE_KEYRING=false
```

## Manual Configuration (Optional)

If you prefer to explicitly disable keyring from the start, or want to avoid the initial keyring attempt, you can use the provided WSL helper script or set the environment variable manually.

### Using the WSL Helper Script

```bash
./run-local-wsl.sh [command] [arguments]
```

### Using Environment Variable

```bash
export NUMEROUS_LOGIN_USE_KEYRING=false
go run main.go [command]
```

## What it does

The CLI credential system now:

1. **Tries keyring first**: Attempts to use the system keyring for credential storage
2. **Detects failures**: Recognizes common keyring unavailability errors (D-Bus issues, missing secret service, etc.)
3. **Automatic fallback**: Seamlessly switches to file-based storage when keyring fails
4. **File-based storage**: Stores credentials in `~/.numerous/credentials.json` with proper permissions
5. **User notification**: Informs you when fallback occurs and provides guidance

## Examples

All standard CLI commands work seamlessly with the new fallback system:

```bash
# Login (will automatically fallback if keyring fails)
go run main.go login

# Check status
go run main.go status

# Deploy an app
go run main.go deploy

# List tasks  
go run main.go task list

# Show help
go run main.go --help
```

## Security Note

When using file-based credentials storage (either manually configured or via automatic fallback), your tokens are stored in `~/.numerous/credentials.json` with `0600` permissions (readable only by you). This file contains sensitive authentication tokens, so:

- **Never commit this file to version control**
- **Ensure your home directory has appropriate permissions**
- **Consider deleting the file when you're done working**

## Environment Variables

### Automatic Mode (Default)
No environment variables needed - the CLI will automatically handle keyring fallback.

### Manual Override
- `NUMEROUS_LOGIN_USE_KEYRING=false` - Forces file-based credential storage from the start

### WSL Script (sets multiple variables)
The WSL script sets these environment variables:
- `NUMEROUS_LOGIN_USE_KEYRING=false` - Enables file-based credential storage
- `NUMEROUS_LOG_LEVEL=debug` - Enables debug logging
- `NUMEROUS_GRAPHQL_HTTP_URL` - GraphQL HTTP endpoint

## Troubleshooting

### Keyring Issues
If you encounter keyring-related errors, the CLI should automatically fall back to file-based storage. Common keyring errors that trigger fallback include:

- `secret service is not available`
- `dbus: session bus is not available` 
- `Cannot autolaunch D-Bus without X11`
- `no keyring available`

### File Permission Issues
If you get permission errors with the credentials file:

```bash
# Fix credentials file permissions
chmod 600 ~/.numerous/credentials.json

# Fix directory permissions  
chmod 700 ~/.numerous/
```

### Migration from Manual to Automatic
If you were previously using manual configuration and want to switch to automatic mode:

1. Remove the `NUMEROUS_LOGIN_USE_KEYRING=false` environment variable
2. Optionally delete the existing credentials file to start fresh: `rm ~/.numerous/credentials.json`
3. Run any CLI command - it will try keyring first and fallback automatically if needed

The automatic fallback system provides the best of both worlds: keyring security when available, and file-based reliability when keyring access fails. 