package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	// Create a simple test file that simulates user input
	testInput := "写一首爵士歌曲\n爵士\n/quit\n"
	
	// Write test input to a file
	os.WriteFile("/tmp/vibe_test_input.txt", []byte(testInput), 0644)
	
	fmt.Println("=== Running TUI test ===")
	fmt.Println("Test input:", testInput)
	
	// Run the TUI with input from file
	cmd := exec.Command("./vibe-echo-new.exe")
	cmd.Stdin, _ = os.Open("/tmp/vibe_test_input.txt")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Println("Starting TUI...")
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting:", err)
		return
	}
	
	// Wait for a bit
	time.Sleep(3 * time.Second)
	
	// Kill if still running
	cmd.Process.Kill()
	cmd.Wait()
	
	fmt.Println("=== Test complete ===")
}
