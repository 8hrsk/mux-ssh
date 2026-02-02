# mux-ssh

mux-ssh is a cross-platform CLI tool designed to manage and connect to SSH servers with ease. It provides a terminal user interface (TUI) for organizing servers and proxies, checking their availability in real-time, and establishing connections quickly.

## Features

- **TUI Dashboard**: A clean, keyboard-navigable interface to view and select servers.
- **Real-time Status Checks**: Automatically checks server availability using SSH handshakes, TCP connections, and ICMP pings.
- **Proxy Support**: Connect to servers via SOCKS5 or HTTP proxies using standard nc (netcat) tunneling.
- **Custom Configuration**: Simple, readable block-based configuration syntax.
- **Cross-Platform**: Works on macOS, Linux (Debian/Ubuntu), and Windows.
- **Editor Integration**: Launch your preferred system or terminal editor directly from the dashboard to manage configurations.

## Installation

### Prerequisites
- **Go 1.21+** (for building from source)
- **OpenSSH Client** (available on most systems)
- **nc** (netcat) for proxy support (standard on macOS/Linux)

### Building from Source
1. Clone the repository:
   ```bash
   git clone https://github.com/8hrsk/mux-ssh.git
   cd mux-ssh
   ```
2. Build the binary:
   ```bash
   go build -o mux-ssh ./cmd/ssh-ogm/main.go
   ```
3. Move the binary to your PATH (optional):
   ```bash
   mv mux-ssh /usr/local/bin/
   ```

## Usage

Run the tool from your terminal:
```bash
mux-ssh
```

### Dashboard Navigation
- **Up/Down (j/k)**: Navigate the list.
- **Left/Right (h/l) or Tab**: Switch between "Servers" and "Proxies" views.
- **Enter**: Connect to the selected server.
- **a**: Add a new server or proxy template to the configuration.
- **r**: Reload configurations and refresh status checks.
- **q**: Quit the application.

### First Run
On the first launch, mux-ssh will create a hidden configuration directory at `~/.ssh-ogm/` containing `config` and `proxies.conf`. You will be prompted to choose your preferred editor (System GUI or Terminal).

## Configuration

Configurations are stored in `~/.ssh-ogm/`.

### Server Configuration (`config`)
Define your SSH servers using the following syntax:

```text
# Production Database
prod-db {
    host: 192.168.1.10
    user: root
    port: 22
    identity: ~/.ssh/id_rsa
}

# Staging Web with Proxy
staging-web {
    host: 10.0.0.5
    user: admin
    proxy: corp-vpn
}
```

**Fields:**
- **host**: IP address or hostname (Required)
- **user**: SSH username (Optional, defaults to current user if omitted by SSH client)
- **port**: SSH port (Optional, defaults to 22)
- **identity**: Path to the private key file (Optional)
- **proxy**: Alias of a proxy defined in `proxies.conf` (Optional)

### Proxy Configuration (`proxies.conf`)
Define proxies to tunnel connections:

```text
# Corporate VPN Proxy
corp-vpn {
    host: vpn.example.com
    port: 1080
    type: socks5
}

# HTTP Gateway
gateway {
    host: proxy.local
    port: 8080
    type: http
}
```

**Fields:**
- **host**: Proxy IP or hostname (Required)
- **port**: Proxy port (Required)
- **type**: Proxy type, either `socks5` or `http` (Required)

## Troubleshooting
- **Connection Failed**: Ensure you have SSH access and the correct keys loaded in your SSH agent.
- **Proxy Issues**: Ensure `nc` is installed and supports the `-x` (proxy) flag.
