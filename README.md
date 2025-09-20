# Zed


Welcome to the **`Zed`** project!

## Usage

The main executable for this project is `Zed`. It requires a TOML configuration file for both the server and client components.


### Detailed Configuration

#### TCP Configuration
* **Server**:

   ```toml
    [mode]
    mode = "server"
    type = "tcp"
    [server]
    address = "127.0.0.1:4000"
    key = "abcdefghijklmnop"
   ```
* **Client**:

   ```toml
    [mode]
    mode = "agent"
    type = "tcp"
    [agent]
    address = ":4000"
    key = "abcdefghijklmnop"
    ports = [
        "8088:8080",
        "9001:9000",
    ]

   ```
* **Details**:

   `address`: The IPv4, IPv6, or domain address of the server to which the client connects.

   `key`: An authentication token used to securely validate and authenticate the connection between the client and server within the tunnel.



#### KCP Configuration
* **Server**:

   ```toml
    [mode]
    mode = "server"
    type = "kcp"
    [kcp]
    ACKNoDelay = false
    Mtu = 1000
    Internal = 1000
    [server]
    address = "127.0.0.1:4000"
    key = "abcdefghijklmnop"
   ```
* **Client**:

   ```toml
    [mode]
    mode = "agent"
    type = "tcp"
    [kcp]
    ACKNoDelay = false
    Mtu = 1000
    Internal = 1000
    [agent]
    address = ":4000"
    key = "abcdefghijklmnop"
    ports = [
        "8088:8080",
        "9001:9000",
    ]

   ```
* **Details**:

   `address`: The IPv4, IPv6, or domain address of the server to which the client connects.

   `key`: An authentication token used to securely validate and authenticate the connection between the client and server within the tunnel.

   `ACKNoDelay`: 
   `Mtu`:
   `Internal`:


## ⚙️ Requirements

- Go ≥ 1.24  
- **Nix** (optional, to get all tools in one shell)


## ❄️ Nix Environment

This project includes a shell.nix file that sets up **Go,golangci** in a reproducible development environment.

```bash
nix-shell
# Build
make build

```

## License

This project is open source and licensed under the **MIT License**.  

You are free to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of this software, **without restriction**, as long as you include the original copyright notice and this permission notice in all copies or substantial portions of the Software.

For the full license text, see the [LICENSE](./LICENSE) file.

