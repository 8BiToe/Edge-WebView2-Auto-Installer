package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	// Microsoft's official WebView2 download URL (evergreen bootstrapper)
	webView2URL   = "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
	installerName = "MicrosoftEdgeWebview2Setup.exe"
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32.NewProc("ShellExecuteW")
)

func main() {
	fmt.Println("Edge WebView2 Auto Installer")
	fmt.Println("===========================")

	// Check if running on Windows
	if runtime.GOOS != "windows" {
		fmt.Println("Error: This program only works on Windows")
		waitAndExit(1)
	}

	// Check if running as administrator
	if !isAdmin() {
		fmt.Println("Administrator privileges required. Requesting elevation...")
		if err := runAsAdmin(); err != nil {
			fmt.Printf("Failed to request admin privileges: %v\n", err)
			fmt.Println("Please run this program as administrator manually.")
			waitAndExit(1)
		}
		// If we reach here, we've launched as admin, so exit this instance
		return
	}

	fmt.Println("Running with administrator privileges ✓")

	// Check if already installed
	if isWebView2Installed() {
		fmt.Println("✓ Edge WebView2 is already installed!")
		fmt.Println("Checking for updates...")
	} else {
		fmt.Println("Edge WebView2 not found. Installing...")
	}

	// Create temp directory for installer
	tempDir, err := os.MkdirTemp("", "webview2_installer")
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		waitAndExit(1)
	}
	defer os.RemoveAll(tempDir)

	installerPath := filepath.Join(tempDir, installerName)

	// Download the installer
	fmt.Println("Downloading Edge WebView2 installer...")
	if err := downloadFile(webView2URL, installerPath); err != nil {
		fmt.Printf("Error downloading installer: %v\n", err)
		waitAndExit(1)
	}
	fmt.Println("✓ Download completed")

	// Run the installer
	fmt.Println("Installing Edge WebView2...")
	if err := runInstaller(installerPath); err != nil {
		fmt.Printf("Error running installer: %v\n", err)
		waitAndExit(1)
	}

	fmt.Println("✓ Edge WebView2 installation completed successfully!")
	fmt.Println("Auto-closing in 3 seconds...")

	// Auto-close after 3 seconds
	time.Sleep(3 * time.Second)
}

// isAdmin checks if the program is running with administrator privileges
func isAdmin() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// runAsAdmin restarts the program with administrator privileges
func runAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Convert strings to UTF16 pointers for Windows API
	verb, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	params, _ := syscall.UTF16PtrFromString("")
	dir, _ := syscall.UTF16PtrFromString("")

	// Call ShellExecuteW to run as administrator
	ret, _, _ := procShellExecute.Call(
		uintptr(0),                      // hwnd
		uintptr(unsafe.Pointer(verb)),   // lpOperation
		uintptr(unsafe.Pointer(exePtr)), // lpFile
		uintptr(unsafe.Pointer(params)), // lpParameters
		uintptr(unsafe.Pointer(dir)),    // lpDirectory
		uintptr(1),                      // nShowCmd (SW_NORMAL)
	)

	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed with code %d", ret)
	}

	return nil
}

// downloadFile downloads a file from URL to filepath
func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// runInstaller executes the WebView2 installer
func runInstaller(installerPath string) error {
	// Run installer with silent install flags
	cmd := exec.Command(installerPath, "/silent", "/install")

	// Set up the command to run with elevated privileges if needed
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	// Run and wait for completion
	if err := cmd.Run(); err != nil {
		// If silent install fails, try interactive install
		fmt.Println("Silent install failed, attempting interactive install...")
		cmd = exec.Command(installerPath)
		return cmd.Run()
	}

	return nil
}

// isWebView2Installed checks if WebView2 is already installed
func isWebView2Installed() bool {
	// Check common installation paths
	paths := []string{
		`C:\Program Files (x86)\Microsoft\EdgeWebView\Application`,
		`C:\Program Files\Microsoft\EdgeWebView\Application`,
		os.ExpandEnv(`${PROGRAMFILES}\Microsoft\EdgeWebView\Application`),
		os.ExpandEnv(`${PROGRAMFILES(X86)}\Microsoft\EdgeWebView\Application`),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// Check for msedgewebview2.exe
			exePath := filepath.Join(path, "msedgewebview2.exe")
			if _, err := os.Stat(exePath); err == nil {
				return true
			}
		}
	}

	// Also check registry (alternative method)
	return checkRegistryForWebView2()
}

// checkRegistryForWebView2 checks Windows registry for WebView2 installation
func checkRegistryForWebView2() bool {
	// Try to run a simple registry query
	cmd := exec.Command("reg", "query",
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
		"/v", "pv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err == nil {
		return true
	}

	// Check 64-bit registry path
	cmd = exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
		"/v", "pv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd.Run() == nil
}

// waitAndExit waits for user input before exiting (only on error)
func waitAndExit(code int) {
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	os.Exit(code)
}
