### GoSystems — High-Performance HTTP/1.1 Server

A from-scratch implementation of an HTTP/1.1 server written in Go.
This project was built to deeply understand how communication on the web actually works — from raw TCP sockets to HTTP request parsing — without relying on external web frameworks.

It focuses on systems-level design, protocol correctness, and memory-efficient performance.

---
#### Key Features:

- **Finite State Machine (FSM) HTTP Parser**

A state-driven parser capable of handling fragmented TCP packets and partial reads correctly.
If a request line or header arrives split across multiple packets, the parser resumes precisely where it left off.

- **Zero-Copy-Oriented Design**

Heavy use of `[]byte` slices and manual buffer management to minimize heap allocations and reduce garbage collector pressure.

- **O(1) Memory Reverse Proxying**

Includes a reverse proxy route that streams responses while computing a SHA-256 checksum
- **Concurrent Connection Handling**

Each client connection is handled independently using goroutines, allowing multiple simultaneous requests without blocking the main listener.


#### 📂 Project Structure
```
HTTPServer/
├── cmd/
│   └── server/
│       └── main.go          # Entry point & routing
├── internal/
│   ├── headers/             # RFC-compliant header parsing & storage
│   ├── request/             # FSM parser & buffer management
│   ├── response/            # HTTP response writer
│   └── server/              # TCP listener & connection handling
├── go.mod                   # Module definition
└── README.md
```
#### Installation & Usage

##### Prerequisites

- Go 1.21+

***Clone the Repository***
```
git clone https://github.com/LD-RW/HTTP-Server.git
cd HTTP-Server
```
***Run the Server***
```
go run cmd/server/main.go
```
***The server starts on:***
```
http://localhost:42069
```
***Test the Streaming Proxy***
```
curl -v http://localhost:42069/httpbin/bytes/100
```
#### Testing

Unit tests are included for core parsing logic to ensure:

- HTTP protocol compliance

- Proper handling of fragmented requests

- Protection against malformed input (e.g., request smuggling attempts)

Run tests with:
```
go test ./internal/...
```
***Note: I built this project to deepen my understanding of how web communication works under the hood of APIs and frameworks.***
