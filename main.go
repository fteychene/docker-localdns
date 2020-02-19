package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/events"
	"github.com/miekg/dns"
)

type ContainerRegistration map[string][]string

var records = ContainerRegistration{}

func sanitizeName(name string) string {
	return strings.TrimPrefix(name, "/")
}

func nameToHostnameDockerLocal(name string) string {
	return sanitizeName(name) + ".docker.local."
}

func nameToHostnameDocker(name string) string {
	return sanitizeName(name) + ".docker."
}

func addToRecords(name string, ip []string) {
	log.Printf("Register record : %s -> %s\n", name, ip)
	records[name] = ip
}

func removeFromRecords(name string) {
	log.Printf("Unregister record : %s\n", name)
	delete(records, name)
}

func containerToService(container types.Container) {
	if container.HostConfig.NetworkMode != "host" {
		for _, name := range container.Names {
			addresableIP := []string{}
			for network, settings := range container.NetworkSettings.Networks {
				if network != "host" {
					addresableIP = append(addresableIP, settings.IPAddress)
				}
			}
			addToRecords(nameToHostnameDockerLocal(name), addresableIP)
			addToRecords(nameToHostnameDocker(name), addresableIP)
			addToRecords(sanitizeName(name)+".", addresableIP)
		}
	}
}

func containerToRegistration(container types.ContainerJSON) {
	if container.HostConfig.NetworkMode != "host" {
		addresableIP := []string{}
		for network, settings := range container.NetworkSettings.Networks {
			if network != "host" {
				addresableIP = append(addresableIP, settings.IPAddress)
			}
		}
		addToRecords(nameToHostnameDockerLocal(container.Name), addresableIP)
		addToRecords(nameToHostnameDocker(container.Name), addresableIP)
		addToRecords(sanitizeName(container.Name)+".", addresableIP)
	}
}

func GenerateRecordsForContainers(client *docker.Client) error {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error calling container list on docker cli")
	}

	for _, container := range containers {
		containerToService(container)
	}
	return nil
}

func handleContainerStart(client *docker.Client, message events.Message) {
	containerInfo, err := client.ContainerInspect(context.Background(), message.Actor.ID)
	if err != nil {
		panic(err)
	}
	containerToRegistration(containerInfo)
}

func handleContainerStop(message events.Message) {
	containerName := message.Actor.Attributes["name"]
	removeFromRecords(nameToHostnameDockerLocal(containerName))
	removeFromRecords(nameToHostnameDocker(containerName))
	removeFromRecords(sanitizeName(containerName) + ".")
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
	dns.HandleFunc(".", handleDNSRequest)
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

	err = GenerateRecordsForContainers(cli)
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
