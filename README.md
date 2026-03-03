Here is a comprehensive, professionally structured README.md tailored specifically for your project. This file highlights your systems-level knowledge and the specific engineering choices we discussed, such as the Finite State Machine and $O(1)$ memory optimizations.GoSystems: High-Performance HTTP/1.1 ServerA from-scratch implementation of an HTTP/1.1 server built in Go. This project explores low-level network programming, the HTTP protocol lifecycle, and high-performance parsing techniques without relying on standard web frameworks.🚀 Key FeaturesFinite State Machine (FSM) Parser: A robust, state-driven parser designed to handle fragmented TCP packets and partial reads gracefully.Zero-Copy Strategy: Heavy reliance on byte slices ([]byte) and manual buffer management to minimize heap allocations and Garbage Collector overhead.Streaming Chunked Encoding: Fully implemented HTTP/1.1 chunked transfer-encoding for dynamic data payloads.$O(1)$ Memory Proxying: A reverse proxy implementation that calculates SHA-256 checksums on streaming data without buffering the entire body into RAM.Concurrent Connection Handling: A non-blocking server engine that leverages Goroutines to handle multiple simultaneous client requests.🛠 Technical Deep Dive1. The State Machine ParserUnlike standard parsers that use complex nested logic, this project utilizes a Finite State Machine (FSM). This ensures that if a network packet is split in the middle of a header or the request line, the server knows exactly where to resume once the next packet arrives.2. Sliding Window Buffer ManagementTo handle data efficiently, the server uses a "Sliding Window" technique in its RequestFromReader function:Read data into a fixed 4KB buffer to prevent unbounded memory usage.Parse consumed bytes and calculate the offset.Shift unconsumed data to the beginning of the buffer for the next read cycle.3. Memory-Efficient HashingThe server includes a specialized route that proxies requests to httpbin.org. Instead of buffering the entire response to calculate a SHA-256 hash (which would consume linear memory), it uses a streaming hash accumulator. This allows the server to process gigabytes of data with a constant memory footprint of only a few kilobytes.📂 Project StructurePlaintextHTTPServer/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point & routing logic
├── internal/
│   ├── headers/             # RFC-compliant header parsing & storage
│   ├── request/             # FSM Parser & buffer management logic
│   ├── response/            # HTTP response formatting & writer
│   └── server/              # TCP listener & concurrency management
├── go.mod                   # Go module definition
└── README.md                # Project documentation
🏗 Installation & UsagePrerequisitesGo 1.21+Getting StartedClone the repository:Bashgit clone https://github.com/LD-RW/HTTP-Server.git
cd HTTPServer
Run the server:Bashgo run cmd/server/main.go
Test the streaming proxy:Bashcurl -v http://localhost:42069/httpbin/bytes/100
🧪 TestingThe project includes unit tests for the core parsing logic to ensure protocol compliance and security (e.g., preventing request smuggling).Bashgo test ./internal/...
Note : I did this to deep my understanding of how the communication on the web happens 
