# Automatic Keyring Fallback

The Numerous CLI now includes automatic fallback functionality for credential storage, making it more reliable across different environments, especially WSL and headless Linux systems.

## How It Works

### Default Behavior (Automatic Mode)

1. **Try Keyring First**: The CLI attempts to use your system's keyring service for secure credential storage
2. **Detect Failures**: If keyring operations fail due to common issues (missing D-Bus, no GUI, etc.), the CLI automatically detects this
3. **Seamless Fallback**: Automatically switches to file-based credential storage without user intervention
4. **User Notification**: Provides helpful feedback about the fallback and how to configure it explicitly if desired

### Common Keyring Errors Detected

The fallback system recognizes these common keyring availability issues:

- `secret service is not available`
- `no keyring available` 
- `keyring not available`
- `dbus: session bus is not available`
- `Cannot autolaunch D-Bus without X11`
- `unknown collection`
- `prompt dismissed`

## Configuration Options

### Automatic Mode (Recommended)

**No configuration required** - just use the CLI normally:

```bash
# Login will automatically handle keyring fallback
numerous login

# All other commands work seamlessly
numerous status
numerous deploy
```

If keyring fails during login, you'll see:
```
Keyring access failed, falling back to file-based credential storage
Using file-based credentials storage: /home/username/.numerous/credentials.json
Note: On WSL/Linux, you can avoid this by setting NUMEROUS_LOGIN_USE_KEYRING=false
```

### Manual File-Based Mode

Force file-based storage from the start:

```bash
export NUMEROUS_LOGIN_USE_KEYRING=false
numerous login
```

This skips the keyring attempt entirely and goes straight to file-based storage.

### WSL Helper Script

For WSL environments, use the provided helper script:

```bash
./run-local-wsl.sh login
./run-local-wsl.sh status
```

This sets up the proper environment variables including `NUMEROUS_LOGIN_USE_KEYRING=false`.

## File-Based Credential Storage

When using file-based storage (either via fallback or explicit configuration):

- **Location**: `~/.numerous/credentials.json`
- **Permissions**: `0600` (readable only by you)
- **Directory Permissions**: `0700` for `~/.numerous/`
- **Format**: JSON with separate fields for access and refresh tokens

### Security Considerations

File-based credential storage is secure for local development:

- ✅ File permissions prevent other users from reading your tokens
- ✅ Tokens are stored in your home directory
- ⚠️ **Never commit credentials.json to version control**
- ⚠️ **Consider deleting the file when done working**

## Migration

### From Manual to Automatic

If you were previously using `NUMEROUS_LOGIN_USE_KEYRING=false`:

1. Remove the environment variable
2. Optionally delete existing credentials: `rm ~/.numerous/credentials.json`  
3. Run any CLI command - it will try keyring first and fallback automatically

### From Keyring to File-Based

If you want to force file-based storage:

1. Set `export NUMEROUS_LOGIN_USE_KEYRING=false`
2. Re-login if needed: `numerous login`

## Troubleshooting

### Keyring Issues

Most keyring issues are now handled automatically. If you continue to have problems:

```bash
# Force file-based storage
export NUMEROUS_LOGIN_USE_KEYRING=false
numerous login
```

### File Permission Issues

Fix credentials file permissions:

```bash
chmod 600 ~/.numerous/credentials.json
chmod 700 ~/.numerous/
```

### Environment-Specific Issues

#### WSL (Windows Subsystem for Linux)
- **Issue**: D-Bus/keyring not available in WSL environment
- **Solution**: Automatic fallback handles this seamlessly, or use explicit file-based mode

#### Docker Containers
- **Issue**: No keyring service in containers
- **Solution**: Automatic fallback works, or set `NUMEROUS_LOGIN_USE_KEYRING=false`

#### Headless Linux Servers
- **Issue**: No GUI/keyring service available
- **Solution**: Automatic fallback works perfectly for CI/CD and server environments

#### macOS
- **Behavior**: Usually works with keychain, fallback available if needed

#### Windows (PowerShell/cmd)
- **Behavior**: Usually works with Windows Credential Manager, fallback available

## Implementation Details

The automatic fallback system:

1. **Error Detection**: Analyzes error messages to identify keyring unavailability
2. **State Management**: Tracks whether fallback has been triggered to avoid repeated attempts
3. **User Experience**: Provides clear feedback about what's happening
4. **Performance**: Only attempts keyring once per session, then uses file-based storage
5. **Compatibility**: Maintains full backward compatibility with existing configurations

## Benefits

- **Zero Configuration**: Works out of the box in all environments
- **Better UX**: No more cryptic keyring errors blocking authentication
- **Cross-Platform**: Consistent behavior across Windows, macOS, Linux, WSL, Docker
- **Development Friendly**: Perfect for CI/CD, containers, and development environments
- **Secure**: File-based storage uses appropriate permissions and follows security best practices

The automatic fallback provides the best of both worlds: keyring security when available, and reliable file-based storage when keyring access fails. 