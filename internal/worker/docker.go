package worker

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Language string

const (
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
)

// Docker container config.
type ContainerConfig struct {
	MemoryLimit string
	CPULimit    string
	Timeout     time.Duration
	CachedDir   string
}

func DockerConfig() ContainerConfig {
	return ContainerConfig{
		MemoryLimit: "128m",
		CPULimit:    "0.5",
		Timeout:     30 * time.Second,
		CachedDir:   "/var/cache/jobrunner",
	}
}

func (c ContainerConfig) BuildDockerArgs(lang Language, code string, rawDeps string) ([]string, error) {
	deps, err := ParseAndSanitizeDeps(rawDeps)
	if err != nil {
		return nil, fmt.Errorf("Dependency error, %w", err)
	}

	args := []string{
		"run",
		"--rm",
		"--memory", c.MemoryLimit,
		"--cpus", c.CPULimit,
	}

	// Determine the language the conatiner will be ran on.
	switch lang {
	case LangPython:
		image := "jobqueue/python-runner:prewarmed"
		cachedVolume := fmt.Sprintf("%s/python:/root/.cached/pip", c.CachedDir)
		script := buildPythonScript(code, deps)

		// Disable the network if there are no extra depenendicies that need to be downloaded.
		if len(deps) == 0 {
			args = append(args, "--network", "none")
		} else {
			// Mount the volume after having downaloded the extra dependencies to prevent redownloading it later on.
			args = append(args, "-v", cachedVolume)
		}

		// Run the script with the dependencies.
		args = append(args, image, "sh", "-c", script)

	case LangJavaScript:
		image := "jobqueue/node-runner:prewarmed"
		cachedVolume := fmt.Sprintf("%s/node:/root/.npm", c.CachedDir)
		script := buildNodeScript(code, deps)

		if len(deps) == 0 {
			args = append(args, "--network", "none")
		} else {
			args = append(args, "-v", cachedVolume)
		}

		args = append(args, image, "sh", "-c", script)
	default:
		return nil, fmt.Errorf("Unsuported language, %s", lang)
	}

	return args, nil
}

// Valid package name regex -- letters, numbers, hyphens, underscores, dots, and version specifiers.
var validPackageRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.\=\<\>\!]+$`)

// Format package names.
func ParseAndSanitizeDeps(raw string) ([]string, error) {
	// Empty string with many spaces.
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	// Seperate dependency names in the string.
	parts := strings.Split(raw, ",")
	clean := make([]string, 0, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}

		if !validPackageRegex.MatchString(trimmed) {
			return nil, fmt.Errorf("Invalid package name format, %q", trimmed)
		}
		clean = append(clean, trimmed)
	}

	return clean, nil
}

func buildPythonScript(code string, deps []string) string {
	// No dependencies.
	if len(deps) == 0 {
		return fmt.Sprintf("python3 -c %q", code)
	}

	// Install the extra dependencies.
	depsList := strings.Join(deps, " ")
	return fmt.Sprintf("pip install --find--links=/root/.cache/pip %s && python3 -c %q", depsList, code)
}

func buildNodeScript(code string, deps []string) string {
	if len(deps) == 0 {
		return fmt.Sprintf("node -e %q", code)
	}

	depsList := strings.Join(deps, " ")
	return fmt.Sprintf("npm install --prefer-offline %s && node -e %q", depsList, code)
}
