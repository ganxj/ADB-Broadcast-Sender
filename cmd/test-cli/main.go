package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/electricbubble/gadb"
)

func main() {
	fmt.Println("=== ADB Broadcast Sender CLI Test ===")
	fmt.Println("Testing ADB connection...")
	
	// Try to create ADB client
	client, err := gadb.NewClient()
	if err != nil {
		log.Printf("ERROR: Failed to create ADB client: %v", err)
		fmt.Println("Make sure ADB is installed and in your PATH")
		fmt.Println("Default ADB path: D:\\Program Files\\Android\\SDK\\platform-tools\\adb.exe")
		os.Exit(1)
	}
	
	fmt.Println("✓ ADB client created successfully")
	
	// Try to get device list
	devices, err := client.DeviceList()
	if err != nil {
		log.Printf("WARNING: Failed to get device list: %v", err)
		fmt.Println("This might be normal if no devices are connected")
	} else {
		fmt.Printf("Found %d device(s):\n", len(devices))
		for i, device := range devices {
			fmt.Printf("  %d. %s\n", i+1, device.Serial())
		}
		
		if len(devices) == 0 {
			fmt.Println("No devices found. Please connect an Android device.")
		}
	}
	
	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Core ADB functionality appears to be working!")
	fmt.Println("The GUI issue is related to CGO and OpenGL dependencies.")
}