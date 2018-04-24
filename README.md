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
- [ ] DNS entry for each network of a container
- [x] Listen docker container start
- [x] Listen docker container stop
- [ ] Listen docker network attach
- [ ] Listen docker network detach
- [ ] Change DNS entries for containers on event
- [ ] Install cobra to create good CLI
- [ ] Configure domain
- [ ] Configure docker socket
- [ ] Systemd service sample
- [ ] Register as dns server for host at startup