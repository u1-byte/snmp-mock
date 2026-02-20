// go-client/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gosnmp/gosnmp"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	// 1. Fetch config from Docker injected .env variables
	targetHost := getEnv("SNMP_TARGET", "127.0.0.1")
	community := getEnv("SNMP_COMMUNITY", "public")
	targetOID := getEnv("SNMP_OID", ".1.3.6.1.2.1.1")

	portStr := getEnv("SNMP_PORT", "161")
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		log.Fatalf("Invalid SNMP_PORT: %v", err)
	}

	// 2. Configure gosnmp
	gosnmp.Default.Target = targetHost
	gosnmp.Default.Port = uint16(port)
	gosnmp.Default.Community = community
	gosnmp.Default.Version = gosnmp.Version2c
	gosnmp.Default.Timeout = time.Duration(2 * time.Second)

	fmt.Printf("Configured Target: %s:%d | Community: %s | Target OID: %s\n",
		targetHost, port, community, targetOID)

	// 3. Connect with a REAL network retry loop
	maxRetries := 10
	connected := false

	for i := 0; i < maxRetries; i++ {
		// Open the local socket
		err = gosnmp.Default.Connect()
		if err == nil {
			// Send a tiny dummy request to test if the remote port is open
			_, err = gosnmp.Default.Get([]string{".1.3.6.1.2.1.1.1.0"})

			if err == nil {
				connected = true
				break
			}
		}

		fmt.Printf("Waiting for %s to start... (Attempt %d/%d) - Err: %v\n", targetHost, i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}

	if !connected {
		log.Fatalf("Server never became available after %d retries.", maxRetries)
	}
	defer gosnmp.Default.Conn.Close()

	fmt.Println("\nSuccessfully verified connection! Starting BulkWalk...\n")

	// 4. Perform BulkWalk and parse specific data types
	err = gosnmp.Default.BulkWalk(targetOID, func(pdu gosnmp.SnmpPDU) error {
		fmt.Printf("OID: %s ", pdu.Name)

		// Parse the value based on the SNMP Data Type
		switch pdu.Type {
		case gosnmp.OctetString:
			b := pdu.Value.([]byte)
			fmt.Printf("| Type: String | Value: %s\n", string(b))
		case gosnmp.Integer:
			val := gosnmp.ToBigInt(pdu.Value)
			fmt.Printf("| Type: Integer | Value: %d\n", val)
		case gosnmp.TimeTicks:
			val := pdu.Value.(uint32)
			fmt.Printf("| Type: TimeTicks | Value: %d\n", val)
		case gosnmp.IPAddress:
			val := pdu.Value.(string)
			fmt.Printf("| Type: IPAddress | Value: %s\n", val)
		default:
			fmt.Printf("| Type: %v | Value: %v\n", pdu.Type, pdu.Value)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("BulkWalk failed: %v", err)
	}

	fmt.Println("\nBulkWalk complete!")
}
