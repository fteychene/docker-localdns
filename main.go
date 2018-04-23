package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/copier"
	"github.com/miekg/dns"
)

type ContainerRegistration map[string][]string

var records = ContainerRegistration{}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for %s\n", q.Name)
			for _, ip := range records[q.Name] {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func nameToHostname(name string) string {
	return strings.TrimPrefix(name, "/") + ".container."
}

func containerToService(container types.Container) ContainerRegistration {
	result := make(ContainerRegistration)
	if container.HostConfig.NetworkMode != "host" {
		for _, name := range container.Names {
			addresableIP := []string{}
			for network, settings := range container.NetworkSettings.Networks {
				if network != "host" {
					addresableIP = append(addresableIP, settings.IPAddress)
				}
			}
			result[nameToHostname(name)] = addresableIP
		}
	}
	return result
}

func MergeContainerRegistration(origin, merge ContainerRegistration) ContainerRegistration {
	result := make(ContainerRegistration)
	copier.Copy(&result, &origin)
	for name, addressableIP := range merge {
		result[name] = addressableIP
	}
	return result
}

func main() {
	// attach request handler func
	dns.HandleFunc("container.", handleDNSRequest)

	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		newRecords := containerToService(container)
		records = MergeContainerRegistration(records, newRecords)
	}

	spew.Dump(records)

	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err = server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
