package apt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	packageNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]+$`)
	versionRegex     = regexp.MustCompile(`^[0-9][a-zA-Z0-9.+:~-]*$`)
)

type DebConfig struct {
	Package       string
	Version       string
	Architecture  string
	Maintainer    string
	Description   string
	Section       string
	Priority      string
	Homepage      string
	InstalledSize int64
	Depends       []string
}

type DebPackageInfo struct {
	Package       string
	Version       string
	Architecture  string
	Maintainer    string
	InstalledSize int64
	Filename      string
	Size          int64
	SHA256        string
	Section       string
	Priority      string
	Homepage      string
	Description   string
}

type RepoConfig struct {
	Origin        string
	Label         string
	Suite         string
	Codename      string
	Architectures []string
	Components    []string
	Description   string
	BaseURL       string
	SigningKeyURL string
}

type ComponentIndex struct {
	Path   string
	Size   int64
	SHA256 string
}

func ValidateDebConfig(cfg DebConfig) error {
	if cfg.Package == "" {
		return errors.New("package name is required")
	}
	if !packageNameRegex.MatchString(cfg.Package) {
		return fmt.Errorf("invalid package name %q: must match %s", cfg.Package, packageNameRegex.String())
	}
	if cfg.Version == "" {
		return errors.New("version is required")
	}
	v := strings.TrimPrefix(cfg.Version, "v")
	if !versionRegex.MatchString(v) {
		return fmt.Errorf("invalid version %q: must start with a digit", cfg.Version)
	}
	if cfg.Architecture == "" {
		return errors.New("architecture is required")
	}
	if cfg.Maintainer == "" {
		return errors.New("maintainer is required")
	}
	if cfg.Description == "" {
		return errors.New("description is required")
	}
	return nil
}

func GenerateControlFile(cfg DebConfig) (string, error) {
	if err := ValidateDebConfig(cfg); err != nil {
		return "", err
	}

	section := cfg.Section
	if section == "" {
		section = "utils"
	}

	priority := cfg.Priority
	if priority == "" {
		priority = "optional"
	}

	v := strings.TrimPrefix(cfg.Version, "v")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Package: %s\n", cfg.Package))
	sb.WriteString(fmt.Sprintf("Version: %s\n", v))
	sb.WriteString(fmt.Sprintf("Section: %s\n", section))
	sb.WriteString(fmt.Sprintf("Priority: %s\n", priority))
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", cfg.Architecture))

	if cfg.InstalledSize > 0 {
		sb.WriteString(fmt.Sprintf("Installed-Size: %d\n", cfg.InstalledSize))
	}

	sb.WriteString(fmt.Sprintf("Maintainer: %s\n", cfg.Maintainer))

	if len(cfg.Depends) > 0 {
		sb.WriteString(fmt.Sprintf("Depends: %s\n", strings.Join(cfg.Depends, ", ")))
	}

	if cfg.Homepage != "" {
		sb.WriteString(fmt.Sprintf("Homepage: %s\n", cfg.Homepage))
	}

	descLines := strings.Split(cfg.Description, "\n")
	sb.WriteString(fmt.Sprintf("Description: %s\n", descLines[0]))
	for i := 1; i < len(descLines); i++ {
		line := descLines[i]
		if strings.TrimSpace(line) == "" {
			sb.WriteString(" .\n")
		} else {
			sb.WriteString(fmt.Sprintf(" %s\n", line))
		}
	}

	return sb.String(), nil
}

func GeneratePackagesIndex(packages []DebPackageInfo) (string, error) {
	if len(packages) == 0 {
		return "", errors.New("package list is empty")
	}

	var blocks []string
	for idx, pkg := range packages {
		if pkg.Package == "" {
			return "", fmt.Errorf("package at index %d has empty name", idx)
		}
		if pkg.Version == "" {
			return "", fmt.Errorf("package at index %d has empty version", idx)
		}
		if pkg.Architecture == "" {
			return "", fmt.Errorf("package at index %d has empty architecture", idx)
		}
		if pkg.Filename == "" {
			return "", fmt.Errorf("package at index %d has empty filename", idx)
		}
		if pkg.Size <= 0 {
			return "", fmt.Errorf("package at index %d has non-positive size", idx)
		}
		if pkg.SHA256 == "" {
			return "", fmt.Errorf("package at index %d has empty sha256 checksum", idx)
		}

		section := pkg.Section
		if section == "" {
			section = "utils"
		}
		priority := pkg.Priority
		if priority == "" {
			priority = "optional"
		}

		v := strings.TrimPrefix(pkg.Version, "v")

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Package: %s\n", pkg.Package))
		sb.WriteString(fmt.Sprintf("Version: %s\n", v))
		sb.WriteString(fmt.Sprintf("Architecture: %s\n", pkg.Architecture))
		sb.WriteString(fmt.Sprintf("Maintainer: %s\n", pkg.Maintainer))
		if pkg.InstalledSize > 0 {
			sb.WriteString(fmt.Sprintf("Installed-Size: %d\n", pkg.InstalledSize))
		}
		sb.WriteString(fmt.Sprintf("Filename: %s\n", pkg.Filename))
		sb.WriteString(fmt.Sprintf("Size: %d\n", pkg.Size))
		sb.WriteString(fmt.Sprintf("SHA256: %s\n", pkg.SHA256))
		sb.WriteString(fmt.Sprintf("Section: %s\n", section))
		sb.WriteString(fmt.Sprintf("Priority: %s\n", priority))
		if pkg.Homepage != "" {
			sb.WriteString(fmt.Sprintf("Homepage: %s\n", pkg.Homepage))
		}
		desc := pkg.Description
		if desc == "" {
			desc = "Wallpaper color scheme and palette converter tool"
		}
		descLines := strings.Split(desc, "\n")
		sb.WriteString(fmt.Sprintf("Description: %s\n", descLines[0]))
		for i := 1; i < len(descLines); i++ {
			line := descLines[i]
			if strings.TrimSpace(line) == "" {
				sb.WriteString(" .\n")
			} else {
				sb.WriteString(fmt.Sprintf(" %s\n", line))
			}
		}

		blocks = append(blocks, sb.String())
	}

	return strings.Join(blocks, "\n") + "\n", nil
}

func GenerateReleaseIndex(cfg RepoConfig, indices []ComponentIndex) (string, error) {
	if cfg.Origin == "" {
		return "", errors.New("origin is required")
	}
	if cfg.Label == "" {
		return "", errors.New("label is required")
	}
	if cfg.Suite == "" {
		return "", errors.New("suite is required")
	}
	if cfg.Codename == "" {
		return "", errors.New("codename is required")
	}
	if len(cfg.Architectures) == 0 {
		return "", errors.New("architectures list cannot be empty")
	}
	if len(cfg.Components) == 0 {
		return "", errors.New("components list cannot be empty")
	}

	dateStr := time.Now().UTC().Format(time.RFC1123Z)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Origin: %s\n", cfg.Origin))
	sb.WriteString(fmt.Sprintf("Label: %s\n", cfg.Label))
	sb.WriteString(fmt.Sprintf("Suite: %s\n", cfg.Suite))
	sb.WriteString(fmt.Sprintf("Codename: %s\n", cfg.Codename))
	sb.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	sb.WriteString(fmt.Sprintf("Architectures: %s\n", strings.Join(cfg.Architectures, " ")))
	sb.WriteString(fmt.Sprintf("Components: %s\n", strings.Join(cfg.Components, " ")))
	if cfg.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", cfg.Description))
	}

	if len(indices) > 0 {
		sb.WriteString("SHA256:\n")
		for _, idx := range indices {
			sb.WriteString(fmt.Sprintf(" %s %d %s\n", idx.SHA256, idx.Size, idx.Path))
		}
	}

	return sb.String(), nil
}

func GenerateSourcesList(cfg RepoConfig) (string, error) {
	if cfg.BaseURL == "" {
		return "", errors.New("baseURL is required")
	}
	suite := cfg.Suite
	if suite == "" {
		suite = "stable"
	}
	components := cfg.Components
	if len(components) == 0 {
		components = []string{"main"}
	}

	if cfg.SigningKeyURL != "" {
		return fmt.Sprintf("deb [signed-by=%s] %s %s %s\n", cfg.SigningKeyURL, cfg.BaseURL, suite, strings.Join(components, " ")), nil
	}
	return fmt.Sprintf("deb %s %s %s\n", cfg.BaseURL, suite, strings.Join(components, " ")), nil
}

func GenerateDeb822Sources(cfg RepoConfig) (string, error) {
	if cfg.BaseURL == "" {
		return "", errors.New("baseURL is required")
	}
	suite := cfg.Suite
	if suite == "" {
		suite = "stable"
	}
	components := cfg.Components
	if len(components) == 0 {
		components = []string{"main"}
	}

	var sb strings.Builder
	sb.WriteString("Types: deb\n")
	sb.WriteString(fmt.Sprintf("URIs: %s\n", cfg.BaseURL))
	sb.WriteString(fmt.Sprintf("Suites: %s\n", suite))
	sb.WriteString(fmt.Sprintf("Components: %s\n", strings.Join(components, " ")))
	if cfg.SigningKeyURL != "" {
		sb.WriteString(fmt.Sprintf("Signed-By: %s\n", cfg.SigningKeyURL))
	}
	if len(cfg.Architectures) > 0 {
		sb.WriteString(fmt.Sprintf("Architectures: %s\n", strings.Join(cfg.Architectures, " ")))
	}
	return sb.String(), nil
}

func ComputeSHA256(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}
