# Docker local dns

# Installation

Compile the project
```bash
make build
```

Start the server (need sudo to expose on port 53) :  
```bash
sudo ./bin/docker-localdns
```

Add the entry `nameserver 127.0.0.1` to `/etc/resolv.conf`

# Usage

```bash
▶ docker run -d nginx
38e23fe936015086ab459b7892172777b4678914ca5333c6db4db7ab291a12dd

▶ docker ps
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS               NAMES
38e23fe93601        nginx               "nginx -g 'daemon of…"   3 seconds ago       Up 1 second         80/tcp              elegant_mcclintock

▶ dig elegant_mcclintock.docker.local
...
;; QUESTION SECTION:
;elegant_mcclintock.docker.local. IN	A

;; ANSWER SECTION:
elegant_mcclintock.docker.local. 3600 IN A	172.17.0.2
...

▶ dig elegant_mcclintock.docker
...
;; QUESTION SECTION:
;elegant_mcclintock.docker. IN	A

;; ANSWER SECTION:
elegant_mcclintock.docker. 3600 IN A	172.17.0.2
...

▶ dig elegant_mcclintock
...
;; QUESTION SECTION:
;elegant_mcclintock. IN	A

;; ANSWER SECTION:
elegant_mcclintock. 3600 IN A	172.17.0.2
...

▶ docker stop elegant_mcclintock
elegant_mcclintock

▶ dig elegant_mcclintock.docker.local
...
;; QUESTION SECTION:
;elegant_mcclintock.docker.local. IN	A

...

▶ docker run --name container1 -d nginx
19fe99c5ae4ddabb9df71d34809a30f667269e73bae3d1d33fd5dd6e0c3e04d4

CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS               NAMES
19fe99c5ae4d        nginx               "nginx -g 'daemon of…"   5 minutes ago       Up 5 minutes        80/tcp              container1


▶ dig container1.docker.local

; <<>> DiG 9.12.1 <<>> container1.docker.local
...
;; QUESTION SECTION:
;container1.docker.local.	IN	A

;; ANSWER SECTION:
container1.docker.local. 3600	IN	A	172.17.0.2
...

```

# Features 

- [x] DNS Server
- [x] Create DNS entries for containers at startup
- [ ] DNS entry for each network of a container
- [x] Listen docker container start/stop/kill/die events
- [ ] Listen docker network attach
- [ ] Listen docker network detach
- [ ] Change DNS entries for containers on event
- [ ] Install cobra to create good CLI
- [ ] Configure domain
- [ ] Configure docker socket
- [ ] Systemd service sample
- [ ] Register as dns server for host at startup
