package main

import (
	"fmt"
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
			cmd = exec.Command("nix-instantiate", "<nixpkgs/nixos>", "-A", "system" ,"-I", "nixos-config="+target)
		} else {
			cmd = exec.Command("nix-instantiate", target)
		}
	}
	output, err  := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to intstantiate nix code: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func IsHashInStore(outputHash string) bool {
	cmd := exec.Command("nix", "store", "path-from-hash-part", outputHash)
	_, err := cmd.Output()
	if err != nil {
		return false
	}
	return true
}
