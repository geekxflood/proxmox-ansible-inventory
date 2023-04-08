package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"gopkg.in/yaml.v2"
)

func main() {
	ipAddr := os.Getenv("PROXMOX_ADDR")
	apiURL := "https://" + ipAddr + ":8006/api2/json"
	username := os.Getenv("PROXMOX_USER")
	password := os.Getenv("PROXMOX_PASS")
	realm := os.Getenv("PROXMOX_REALM")

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	client, err := proxmox.NewClient(apiURL, httpClient, "", nil, "", 0)
	if err != nil {
		log.Fatal("Error connecting to Proxmox:", err)
	}

	err = client.Login(username+"@"+realm, password, "")
	if err != nil {
		log.Fatal("Error logging in to Proxmox:", err)
	}

	vmList, err := client.GetVmList()
	if err != nil {
		log.Fatal("Error retrieving VM list:", err)
	}

	inventory := make(map[string]interface{})
	inventory["all"] = make(map[string]interface{})
	hosts := make(map[string]interface{})

	for _, vm := range vmList["data"].([]interface{}) {
		vmMap := vm.(map[string]interface{})
		template := vmMap["template"].(float64)

		if template == 0 {
			hostname := vmMap["name"].(string)
			vmid := int(vmMap["vmid"].(float64))
			vmr := proxmox.NewVmRef(vmid)

			interfaces, err := client.GetVmAgentNetworkInterfaces(vmr)
			if err == nil {
				for _, iface := range interfaces {
					if iface.Name == "eth0" && len(iface.IPAddresses) > 0 {
						hostVars := make(map[string]interface{})
						hostVars["ansible_host"] = iface.IPAddresses[0].String()
						hosts[hostname] = hostVars
						break
					}
				}
			}
		}
	}

	inventory["all"].(map[string]interface{})["hosts"] = hosts

	inventoryYaml, err := yaml.Marshal(inventory)
	if err != nil {
		log.Fatal("Error generating Ansible inventory YAML:", err)
	}

	fmt.Println(string(inventoryYaml))

	err = os.WriteFile("inventory.yaml", inventoryYaml, 0644)
	if err != nil {
		log.Fatal("Error writing inventory.yaml file:", err)
	}
}
