package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alexflint/go-filemutex"
	"github.com/mpdroog/transip"
)

func getTransIPClient(username, privKeyPath string) (*transip.Client, error) {
	client := &transip.Client{
		Login:     username,
		ReadWrite: true,
	}

	if err := client.SetPrivateKeyFromPath(privKeyPath); err != nil {
		return nil, fmt.Errorf("Could not load private key from path %s: %s", privKeyPath, err)
	}

	return client, nil
}

func main() {
	args := os.Args[1:]
	log.Printf("TransIP DNS provider: Add challenge DNS entry TXT %s %s (TTL %s)", args[1], args[2], args[3])

	// We need a lockfile, since Traefik will just invoke this multiple times for different ACME challenges
	// and we don't want them to get in the way of eachother
	mutex, err := filemutex.New("/tmp/transip.lock")
	if err != nil {
		log.Fatal("Lockfile directory did not exist or file could not be created")
	}

	log.Println("Waiting for lockfile")
	mutex.Lock()

	// Create the TransIP API client
	client, err := getTransIPClient(os.Getenv("TRANSIP_USERNAME"), os.Getenv("TRANSIP_PRIVATE_KEY_PATH"))
	if err != nil {
		mutex.Unlock()
		log.Fatal("Error", err)
	}

	// ACME challenge domains end with a .
	parts := strings.Split(args[1], ".")
	rootDomain := strings.Join([]string{parts[len(parts)-3], parts[len(parts)-2]}, ".")
	newEntryName := strings.Join(parts[0:len(parts)-3], ".")

	// Create TransIP Domain Service
	domainService := transip.DomainService{
		Creds: *client,
	}

	// Retrieve current DNS entries
	domain, err := domainService.Domain(rootDomain)
	if err != nil {
		mutex.Unlock()
		log.Fatal("Error retrieving DNS entries", err)
	}

	// Print the current DNS entries
	log.Println("Current DNS entries:")
	log.Println("\tName Expire Type Content")
	for _, entry := range domain.DNSEntry {
		log.Printf("\t%s %d %s %s", entry.Name, entry.Expire, entry.Type, entry.Content)
	}

	// Create the new DNS entry
	ttl, err := strconv.Atoi(args[3])
	ttlOverride := os.Getenv("OVERRIDE_DNS_TTL")
	if ttlOverride != "" {
		log.Printf("Using override TTL to set DNS record: %s", ttlOverride)
		ttl, err = strconv.Atoi(ttlOverride)
	}
	if err != nil {
		mutex.Unlock()
		log.Fatal("Unable to convert TTL to integer")
	}
	newEntry := transip.DomainDNSentry{
		Name:    newEntryName,
		Expire:  ttl,
		Type:    "TXT",
		Content: args[2],
	}

	// Check for existing entries that need to be removed
	existingIndex := -1
	entries := domain.DNSEntry
	for idx, entry := range domain.DNSEntry {
		if entry.Name == newEntry.Name && entry.Type == newEntry.Type {
			existingIndex = idx
		}
	}
	if existingIndex > -1 {
		log.Println("Removing old challenge entry")
		entries = append(entries[:existingIndex], entries[existingIndex+1:]...)
	}

	// Append the new entry to the list, and send the list to TransIP
	entries = append(entries, newEntry)

	log.Println("Writing DNS entries")
	err = domainService.SetDNSEntries(rootDomain, entries)
	if err != nil {
		mutex.Unlock()
		log.Fatal("Error setting new DNS entries", err)
	}

	log.Println("DNS entries set")
	mutex.Unlock()
}
