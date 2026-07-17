# Zoho SETU Secure File Server

A highly concurrent, cross-platform, secure enterprise file server built in Go.

## Features
- **Cross-Platform:** Native binaries for Windows, macOS, and Linux.
- **Server Administrator GUI:** Built with the Fyne toolkit, allowing administrators to manage network configurations and user accounts locally.
- **Embedded Web Client:** A responsive, dark-themed Web UI for end-users to upload, download, and manage files from their browsers or mobile devices.
- **Virtual File System (VFS):** Abstracted storage layer ensuring strict separation of concerns and allowing for future cloud integrations (e.g., AWS S3).
- **Thread-Safe Concurrency:** Granular `sync.RWMutex` file locking prevents data corruption during simultaneous multi-user uploads and downloads.
- **Robust Security:** 
  - Automated self-signed ECC TLS certificate generation (HTTPS).
  - Passwords hashed securely in memory using `bcrypt`.
  - Comprehensive HTTP audit logging middleware.

## Getting Started

### Prerequisites
- [Go](https://golang.org/doc/install) installed on your system.
- C compiler (GCC/Clang) required for Fyne GUI dependencies.

### Running the Server
Clone the repository and start the Administrator GUI:
```bash
go mod tidy
go run ./cmd/server-gui
```
1. Configure your port (default `8080`) and click **Start Server**.
2. Navigate to the **User Management** tab to provision a secure user account.

### Accessing the Web Client
Once the server is running, open a web browser on your computer or any device on your local network:
```
https://localhost:8080
```
*(Note: Because the server generates self-signed TLS certificates for local security, your browser will display a "Not Private" warning. You must bypass this warning to access the encrypted connection).*

## Engineering Documentation
For a deep dive into the architectural decisions, concurrency model, and system design, compile and read the included `Zoho_SETU_Writeup.tex` research paper.

## License
MIT
