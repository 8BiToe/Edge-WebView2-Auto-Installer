# Edge WebView2 Auto Installer

This utility automatically installs or repairs Microsoft Edge WebView2 on Windows systems, addressing issues where Microsoft programs fail to update or install WebView2 due to misconfigured permissions or missing components.

## Why Use This Tool?
Some Microsoft and third-party applications require Edge WebView2 to function correctly. Occasionally, Windows systems encounter problems where WebView2 cannot be updated or installed, often due to insufficient permissions or corrupted installations. This tool:
- Checks if Edge WebView2 is installed and up-to-date
- Requests administrator privileges if needed
- Downloads the official Microsoft WebView2 installer
- Runs the installer silently (or interactively if silent install fails)
- Cleans up temporary files after installation

## How It Works
1. **Platform Check:** Ensures the program is running on Windows.
2. **Permission Elevation:** Requests administrator rights if not already running as admin.
3. **Installation Check:** Looks for existing WebView2 installations in common locations and the Windows registry.
4. **Download & Install:** Downloads the latest official WebView2 installer from Microsoft and runs it.
5. **Feedback:** Provides clear console output for each step and auto-closes after completion.

## Usage
Simply run the executable (`webview_installer.exe`). If administrator privileges are required, the program will prompt for elevation. No manual downloads or command-line arguments are needed.

## When to Use
- If Microsoft or other programs complain about missing or outdated Edge WebView2
- If WebView2 installation fails due to permission errors
- For IT support or automated repair scripts

## Notes
- This tool only works on Windows.
- Requires an internet connection to download the official installer.

---

*Created to simplify and automate the repair of Edge WebView2 installations on Windows systems.*