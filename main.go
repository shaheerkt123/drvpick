package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func GetDrvPath(target string, isFlake bool, isSystem bool) (string, error) {
	var cmd *exec.Cmd
	if isFlake {
		if isSystem {
			cmd = exec.Command("nix", "eval", "--raw", target+".config.system.build.toplevel.drvPath")
		} else {
			cmd = exec.Command("nix", "eval", "--raw", target+".drvPath")
		}
	} else {
		if isSystem {
			cmd = exec.Command("nix-instantiate", "<nixpkgs/nixos>", "-A", "system", "-I", "nixos-config="+target)
		} else {
			cmd = exec.Command("nix-instantiate", "-E", "with import <nixpkgs> {}; callPackage "+target+" {}")
		}
	}
	output, err  := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to intstantiate nix code: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func IsOutputInStore(outPath string) bool {
	cmd := exec.Command("nix", "path-info", "/nix/store/" + outPath)
	_, err := cmd.Output()
	if err != nil {
		return false
	}
	return true
}

func IsOutputInCaches(outPath string, caches string, publicKeys string) (string, error) {
	cmd := exec.Command("nix", "build", "--dry-run", "/nix/store/" + outPath,
	"--option", "substituters", caches,
	"--option", "trusted-public-keys", publicKeys)
	
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("error creating pipe: %w", err)
	}

	// You must start the command before reading from the pipe
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %w", err)
	}

	outputBytes, err := io.ReadAll(pipe)
	if err != nil {
		return "", fmt.Errorf("error reading command output: %w", err)
	}

	// Wait for the command to finish execution
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("command finished with error: %w", err)
	}

	outputString := string(outputBytes)

	start := strings.Index(outputString, "(")
	end := strings.Index(outputString, ")")

	if start != -1 && end > start {
		return outputString[start:end+1], nil
	} else {
		return "", fmt.Errorf("not in the caches")
	}
}

func BuildDerivation(drvPath string, specificOutput string, caches string, publicKeys string) error {
	cmd := exec.Command("nix", "build", drvPath + "^" + specificOutput,
	"--option", "subtituters", caches,
	"--option", "trusted-public-keys", publicKeys)

	_, err := cmd.Output()
	return err
} 

