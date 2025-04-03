package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	// "encoding/json"
)

type SystemInfo struct {
	OS         string            `json: "os"`
	pkgManager string            `json: "pkgManager"`
	Packages   map[string]string `json: "packages"`
	Snaps      map[string]string `json: "snaps"`
	Flatpaks   map[string]string `json: "flatpaks"`
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

func FetchSystemInfo() {
	os := DetectSystem()
	pkgManager := DetectPkgManager()
	packages := ListPackages(pkgManager)
	configs, _ := FindConfigFiles(packages)

	fmt.Printf("%s, %s \n", os, pkgManager)

	for pkg, conf := range configs {
		fmt.Printf("%s: %s\n", pkg, conf)
	}
	fmt.Println(len(packages))
	fmt.Println(len(configs))
}
