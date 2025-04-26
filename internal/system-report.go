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
	OS         string              `json:"os"`
	PkgManager string              `json:"pkgManager"`
	Packages   []string            `json:"packages"`
	Configs    map[string][]string `json:"configs"`
	Snaps      []string            `json:"snaps"`
	Flatpaks   []string            `json:"flatpaks"`
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
		cmd = exec.Command("cat", "/etc/apk/world")
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

func ListSnapPackages() []string {
	var snaps []string
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
			snaps = append(snaps, fields[0])
		}
	}
	return snaps
}

func ListFlatpakPackages() []string {
	var flatpaks []string
	cmd := exec.Command("flatpak", "list", "--app")
	output, err := cmd.Output()
	if err != nil {
		return flatpaks
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			flatpaks = append(flatpaks, fields[0])
		}
	}
	return flatpaks
}

func normalize(name string) string {
	name = strings.ToLower(name)
	suffixes := []string{"-git", "-bin", "-stable", "-nightly", "-dev", "-browser"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

func FindConfigFiles(pkgs []string) (map[string][]string, error) {
	results := make(map[string][]string)
	total := len(pkgs)
	home := os.Getenv("HOME")

	configLocations := []string{
		filepath.Join(home, ".config"),
		//filepath.Join(home, ".local", "share"),
		//filepath.Join(home, ".local", "config"),
		//filepath.Join(home, ".cache"),
		filepath.Join(home, ".var", "app"),
	}

	for i, pkg := range pkgs {
		pkgNorm := normalize(pkg)
		for _, location := range configLocations {
			configNorm := filepath.Join(location, pkgNorm)
			config := filepath.Join(location, pkg)

			if _, err := os.Stat(configNorm); err == nil {
				results[pkg] = append(results[pkg], configNorm)
			} else if _, err := os.Stat(config); err == nil {
				results[pkg] = append(results[pkg], config)
			}

		}

		dotPkg := "." + pkgNorm
		dotPkgPath := filepath.Join(home, dotPkg)
		if _, err := os.Stat(dotPkgPath); err == nil {
			results[pkg] = append(results[pkg], dotPkgPath)
		}

		if results[pkg] == nil {
			results[pkg] = append(results[pkg], "-")
		}

		fmt.Printf("\rProgress: %d/%d", i+1, total)
	}

	fmt.Println()
	return results, nil
}

func GenerateJSONFile(fileName string, system SystemInfo) {
	var data any

	switch fileName {
	case "packages":
		data = system.Packages
	case "configs":
		data = system.Configs
	case "system":
		data = system
	case "snaps":
		data = system.Snaps
	case "flatpaks":
		data = system.Configs
	}

	fileName = fileName + ".json"

	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println("Error encoding JSON")
	}

	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing system info to JSON file")
	}
}

func copyWithCP(src, dest string) error {
	cmd := exec.Command("cp", "-r", src, dest)
	return cmd.Run()
}

func CollectConfigs(configs map[string][]string, directory string) error {
	for pkg, paths := range configs {
		pkg = normalize(pkg)
		for _, path := range paths {
			if path == "-" {
				continue
			}

			destPath := filepath.Join(directory, pkg)

			err := os.MkdirAll(destPath, 0755)
			if err != nil {
				fmt.Printf("Failed to create destination %s: %v\n", destPath, err)
				continue
			}
			err = copyWithCP(path, destPath)
			if err != nil {
				fmt.Printf("Couldn't copy %s to %s: %v\n", path, destPath, err)
				continue
			}
		}
	}

	return nil
}

func Backup() {
	baseDir := filepath.Join(os.Getenv("HOME"), "go-figure")
	err := os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating go-figure directory")
	}

	distro := DetectSystem()
	pkgManager := DetectPkgManager()
	packages := ListPackages(pkgManager)
	configs, _ := FindConfigFiles(packages)
	snaps := ListSnapPackages()
	flatpaks := ListFlatpakPackages()

	system := SystemInfo{
		OS:         distro,
		PkgManager: pkgManager,
		Packages:   packages,
		Configs:    configs,
		Snaps:      snaps,
		Flatpaks:   flatpaks,
	}

	CollectConfigs(configs, baseDir)

	GenerateJSONFile("system", system)
	GenerateJSONFile("packages", system)
	GenerateJSONFile("configs", system)
}
