package helpers

import (
	"os/exec"
)

// ExecutePowershellCommand runs a PowerShell command and returns the output or an error
func ExecutePowershellCommand(powershellbinary string, command string) (string, error) {
	cmd := exec.Command(powershellbinary, "-Command", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
