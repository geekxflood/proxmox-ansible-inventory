package main

import (
	"log"
	"os"
)

func main() {
	log.Println("Starting Proxmox Ansible Inventory Generator...")

	ipAddr := os.Getenv("PROXMOX_ADDR")
	username := os.Getenv("PROXMOX_USER")
	password := os.Getenv("PROXMOX_PASS")
	realm := os.Getenv("PROXMOX_REALM")

	client := createProxmoxClient(ipAddr)
	authenticateProxmoxClient(client, username, password, realm)
	vmList := fetchVmList(client)

	inventory := createInventory(client, vmList)

	if len(inventory.Data["all"].(map[string]interface{})["hosts"].(map[string]interface{})) == 0 {
		log.Fatal("No hosts found in inventory")
	} else {
		inventoryYaml := convertInventoryToYaml(inventory)
		writeInventoryToFile(inventoryYaml)
	}
}
