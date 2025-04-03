package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SystemInfo struct {
	OS         string            `json:"os"`
	PkgManager string            `json:"pkgManager"`
	Packages   map[string]string `json:"packages"`
	Snaps      map[string]string `json:"snaps"`
	Flatpaks   map[string]string `json:"flatpaks"`
}

func DetectSystem() string {
	if data, err := os.ReadFile("/etc/os-release"); err != nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ID=") {
				osName := strings.Trim(line, "ID=")
				osName = strings.Trim(osName, `"`)
				return osName
			}
		}
	}

	// Fallback
	cmd := exec.Command("lsb_release", "-si")
	output, err := cmd.Output()
	if err == nil {
		return strings.ToLower(strings.TrimSpace(string(output)))
	}

	return "unknown"
}

func DetectPkgManager() string {
	pmCommands := [8]string{
		"apt",
		"dnf",
		"yum",
		"pacman",
		"apk",
		"zypper",
		"portage",
		"xbps",
	}

	for _, cmd := range pmCommands {
		_, err := exec.Command(cmd, "--version").CombinedOutput()
		if err == nil {
			return cmd
		}
	}

	return "unknown"
}

func ListPackages(pkgManager string) []string {
	var cmd *exec.Cmd

	switch pkgManager {
	case "apt":
		cmd = exec.Command("apt-mark", "showmanual")
	case "dnf":
		cmd = exec.Command("dnf", "repoquery", "--userinstalled", "--qf", "%{NAME}")
	case "pacman":
		cmd = exec.Command("pacman", "-Qe", "--quiet")
	case "apk":
		cmd = exec.Command("apk", "info", "--installed")
	case "zypper":
		cmd = exec.Command("zypper", "se", "--installed-only")
	case "portage":
		cmd = exec.Command("qlist", "-Ive")
	case "xbps":
		cmd = exec.Command("xbps-query", "-m")
	default:
		return []string{}
	}

	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	var packages []string
	for _, line := range lines {
		pkg := strings.TrimSpace(line)
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}
	return packages
}

func ListSnapPackages() map[string]string {
	snaps := make(map[string]string)

	cmd := exec.Command("snap", "list")
	output, err := cmd.Output()
	if err != nil {
		return snaps
	}

	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 0 {
			snaps[fields[0]] = "installed"
		}
	}
	return snaps
}

func ListFlatpakPackages() map[string]string {
	flatpaks := make(map[string]string)

	cmd := exec.Command("flatpak", "list", "--app")
	output, err := cmd.Output()
	if err != nil {
		return flatpaks
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			flatpaks[fields[0]] = "installed"
		}
	}
	return flatpaks
}

func FindConfigFiles(pkgs []string) (map[string]string, error) {
	results := make(map[string]string)
	total := len(pkgs)

	for i, pkg := range pkgs {
		found := false
		var foundPath string

		searchDirs := []string{"/etc", "/var", os.Getenv("HOME") + "/.config"}
		for _, dir := range searchDirs {
			var matches []string
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				name := strings.ToLower(info.Name())
				if strings.Contains(name, strings.ToLower(pkg)) {
					matches = append(matches, path)
					return filepath.SkipDir
				}
				return nil
			})

			if err == nil && len(matches) > 0 {
				foundPath = matches[0]
				found = true
				break
			}
		}

		if !found {
			additionalDirs := []string{"/etc", "/etc/default", "/etc/sysconfig"}
			for _, dir := range additionalDirs {
				potentialPath := filepath.Join(dir, pkg)
				if _, err := os.Stat(potentialPath); err == nil {
					foundPath = potentialPath
					found = true
					break
				}
			}
		}

		if found {
			results[pkg] = foundPath
		} else {
			results[pkg] = "Not found"
		}

		fmt.Printf("\rProgress: %d/%d", i+1, total)
	}

	fmt.Println()
	return results, nil
}

func Backup() {
	distro := DetectSystem()
	pkgManager := DetectPkgManager()
	packages := ListPackages(pkgManager)
	configs, _ := FindConfigFiles(packages)
	snaps := ListSnapPackages()
	flatpaks := ListFlatpakPackages()

	system := SystemInfo{
		OS:         distro,
		PkgManager: pkgManager,
		Packages:   configs,
		Snaps:      snaps,
		Flatpaks:   flatpaks,
	}

	jsonData, err := json.MarshalIndent(system, "", " ")
	if err != nil {
		fmt.Println("Error encoding JSON")
	}

	fileName := "backup.json"
	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing system info to JSON file")
	}

	fmt.Printf("%s, %s \n", system.OS, system.PkgManager)
	fmt.Println("Packages:", len(system.Packages))
	fmt.Println("Snaps:", len(system.Snaps))
	fmt.Println("Flatpaks:", len(system.Flatpaks))
}
