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

func normalize(name string) string {
	name = strings.ToLower(name)
	suffixes := []string{"-git", "-bin", "-stable", "-nightly", "-dev", "-browser"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

func FindConfigFiles(pkgs []string) (map[string]string, error) {
	results := make(map[string]string)
	total := len(pkgs)
	home := os.Getenv("HOME")

	configLocations := []string{
		filepath.Join(home, ".config"),
		home,
		filepath.Join(home, ".local", "share"),
		filepath.Join(home, ".local", "config"),
		filepath.Join(home, ".cache"),
		filepath.Join(home, ".var", "app"),
	}

	for i, pkg := range pkgs {
		found := false
		pkgNorm := normalize(pkg)
		entries, err := os.ReadDir(configLocations[0])
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					dirNorm := normalize(entry.Name())
					if pkgNorm == dirNorm {
						found = true
						results[pkg] = filepath.Join(configLocations[0], entry.Name())
						break
					}
				}
			}
		}

		if !found {
			dotPkg := "." + pkgNorm
			dotPkgPath := filepath.Join(home, dotPkg)
			if _, err := os.Stat(dotPkgPath); err == nil {
				found = true
				results[pkg] = dotPkgPath
			}
		}

		if !found {
			for j := 2; j < len(configLocations); j++ {
				configDir := configLocations[j]
				if _, err := os.Stat(configDir); os.IsNotExist(err) {
					continue
				}

				entries, err := os.ReadDir(configDir)
				if err != nil {
					continue
				}

				for _, entry := range entries {
					name := entry.Name()
					if normalize(name) == pkgNorm {
						found = true
						results[pkg] = filepath.Join(configDir, name)
						break
					}
				}

				if found {
					break
				}
			}
		}

		if !found {
			results[pkg] = "-"
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
