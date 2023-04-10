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

func createProxmoxClient(ipAddr string) *proxmox.Client {
	apiURL := fmt.Sprintf("https://%s:8006/api2/json", ipAddr)

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

	return client
}

func authenticateProxmoxClient(client *proxmox.Client, username, password, realm string) {
	err := client.Login(fmt.Sprintf("%s@%s", username, realm), password, "")
	if err != nil {
		log.Fatal("Error logging in to Proxmox:", err)
	}
}

func fetchVmList(client *proxmox.Client) map[string]interface{} {
	vmList, err := client.GetVmList()
	if err != nil {
		log.Fatal("Error retrieving VM list:", err)
	}

	return vmList
}

func createInventory(client *proxmox.Client, vmList map[string]interface{}) Inventory {
	inventory := Inventory{Data: make(map[string]interface{})}
	inventory.Data["all"] = make(map[string]interface{})
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

	inventory.Data["all"].(map[string]interface{})["hosts"] = hosts
	return inventory
}

func convertInventoryToYaml(inventory Inventory) []byte {
	inventoryYaml, err := yaml.Marshal(inventory.Data)
	if err != nil {
		log.Fatal("Error generating Ansible inventory YAML:", err)
	}

	return inventoryYaml
}

func writeInventoryToFile(inventoryYaml []byte) {
	outputDir := "output"
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatal("Error creating output directory:", err)
	}

	outputFile := fmt.Sprintf("%s/inventory.yaml", outputDir)
	err = os.WriteFile(outputFile, inventoryYaml, 0644)
	if err != nil {
		log.Fatal("Error writing inventory.yaml file:", err)
	}

	log.Printf("Ansible inventory written to %s", outputFile)
}
