# Docker local dns

# Installation

TODO

# Usage

Compile the project
```bash
make build
```

Start the server (need sudo to expose on port 53) :  
```bash
sudo ./bin/docker-localdns
```

Add the entry `nameserver 127.0.0.1` to `/etc/resolv.conf`

# Features 

- [x] DNS Server
- [x] Create DNS entries for containers at startup
- [ ] Listen docker event
- [ ] Change DNS entries for containers on event
- [ ] Install cobra to create good CLI
- [ ] Configure domain
- [ ] Configure docker socket
- [ ] Systemd service sample
- [ ] Register as dns server for host at startup