package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/events"
	"github.com/jinzhu/copier"
	"github.com/miekg/dns"
)

type ContainerRegistration map[string][]string

var records = ContainerRegistration{}

func nameToHostname(name string) string {
	return strings.TrimPrefix(name, "/") + ".docker.local."
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

func containerToRegistration(container types.ContainerJSON) ContainerRegistration {
	result := make(ContainerRegistration)
	if container.HostConfig.NetworkMode != "host" {
		addresableIP := []string{}
		for network, settings := range container.NetworkSettings.Networks {
			if network != "host" {
				addresableIP = append(addresableIP, settings.IPAddress)
			}
		}
		result[nameToHostname(container.Name)] = addresableIP
	}
	return result
}

func GenerateRecordsForContainers(client *docker.Client) (ContainerRegistration, error) {
	result := make(ContainerRegistration)
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Error calling container list on docker cli")
	}

	for _, container := range containers {
		newRecords := containerToService(container)
		result = MergeContainerRegistration(result, newRecords)
	}
	return result, nil
}

func handleContainerStart(client *docker.Client, message events.Message) {
	containerInfo, err := client.ContainerInspect(context.Background(), message.Actor.ID)
	if err != nil {
		panic(err)
	}
	newRecords := containerToRegistration(containerInfo)
	records = MergeContainerRegistration(records, newRecords)
	spew.Dump(records)
}

func handleContainerStop(message events.Message) {
	delete(records, nameToHostname(message.Actor.Attributes["name"]))
	spew.Dump(records)
}

func ListenToDockerEvents(client *docker.Client) error {
	events, _ := client.Events(context.Background(), types.EventsOptions{})
	for event := range events {
		if event.Type == "container" && event.Action == "start" {
			handleContainerStart(client, event)
		} else if event.Type == "container" && (event.Action == "die" || event.Action == "stop" || event.Action == "kill") {
			handleContainerStop(event)
		}
	}
	return nil
}

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

func StartDNSServer(host string, port int) (*dns.Server, error) {
	dns.HandleFunc("docker.local.", handleDNSRequest)
	server := &dns.Server{Addr: host + ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	if err != nil {
		return nil, errors.Wrap(err, "Error starting dns server")
	}
	return server, nil
}

func main() {
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	records, err = GenerateRecordsForContainers(cli)
	if err != nil {
		log.Fatalf("%e\n", err)
		panic(err)
	}
	go ListenToDockerEvents(cli)

	// start server
	port := 53

	dnsServer, err := StartDNSServer("", port)
	defer dnsServer.Shutdown()
	if err != nil {
		log.Fatalf("%e\n ", err)
		panic(err)
	}
}
