package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func setLED(muted bool) {
	value := "0"

	if muted {
		value = "1"
	}

	cmd := exec.Command(
		"sudo",
		"/usr/local/bin/micled-set",
		value,
	)

	err := cmd.Run()
	if err != nil {
		fmt.Println("LED error:", err)
		return
	}

	fmt.Println("LED set:", value)
}

func isMicMuted() bool {
	cmd := exec.Command(
		"pactl",
		"list",
		"sources",
	)

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("pactl list error:", err)
		return false
	}

	text := string(out)

	for _, block := range strings.Split(text, "Source #") {
		if strings.Contains(block, "Name: alsa_input.pci-0000_00_1f.3.analog-stereo") {
			return strings.Contains(block, "Mute: yes")
		}
	}

	fmt.Println("microphone source not found")
	return false
}

func syncLED() {
	muted := isMicMuted()

	fmt.Println("microphone muted:", muted)

	setLED(muted)
}

func main() {
	fmt.Println("ThinkPad mic LED daemon started")

	fmt.Println("XDG_RUNTIME_DIR:", os.Getenv("XDG_RUNTIME_DIR"))
	fmt.Println("USER:", os.Getenv("USER"))

	syncLED()

	cmd := exec.Command("pactl", "subscribe")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Println("pactl error:", scanner.Text())
		}
	}()

	fmt.Println("waiting for PipeWire events...")

	reader := bufio.NewScanner(stdout)

	for reader.Scan() {
		line := reader.Text()

		fmt.Println("event:", line)

		if strings.Contains(line, "source") {
			time.Sleep(100 * time.Millisecond)
			syncLED()
		}
	}

	err = cmd.Wait()

	fmt.Println("pactl exited:", err)
}
