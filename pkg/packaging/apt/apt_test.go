package apt

import (
	"strings"
	"testing"
)

func TestValidateDebConfig_Valid(t *testing.T) {
	cfg := DebConfig{
		Package:      "gowall",
		Version:      "v0.2.4",
		Architecture: "amd64",
		Maintainer:   "Adam Magued <adamismailmageud@gmail.com>",
		Description:  "Wallpaper color scheme and palette converter tool",
	}

	if err := ValidateDebConfig(cfg); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateDebConfig_InvalidPackageName(t *testing.T) {
	tests := []struct {
		name    string
		pkgName string
	}{
		{"empty", ""},
		{"uppercase", "Gowall"},
		{"underscore", "go_wall"},
		{"starts with hyphen", "-gowall"},
		{"invalid symbols", "gowall@tool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DebConfig{
				Package:      tt.pkgName,
				Version:      "0.2.4",
				Architecture: "amd64",
				Maintainer:   "Maintainer",
				Description:  "Description",
			}
			if err := ValidateDebConfig(cfg); err == nil {
				t.Fatalf("expected error for invalid package name %q, got nil", tt.pkgName)
			}
		})
	}
}

func TestValidateDebConfig_InvalidVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"empty", ""},
		{"starts with letter without v", "alpha-0.1"},
		{"special character start", "-1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DebConfig{
				Package:      "gowall",
				Version:      tt.version,
				Architecture: "amd64",
				Maintainer:   "Maintainer",
				Description:  "Description",
			}
			if err := ValidateDebConfig(cfg); err == nil {
				t.Fatalf("expected error for invalid version %q, got nil", tt.version)
			}
		})
	}
}

func TestValidateDebConfig_MissingRequiredFields(t *testing.T) {
	base := DebConfig{
		Package:      "gowall",
		Version:      "0.2.4",
		Architecture: "amd64",
		Maintainer:   "Maintainer",
		Description:  "Description",
	}

	t.Run("missing architecture", func(t *testing.T) {
		cfg := base
		cfg.Architecture = ""
		if err := ValidateDebConfig(cfg); err == nil {
			t.Fatal("expected error for missing architecture, got nil")
		}
	})

	t.Run("missing maintainer", func(t *testing.T) {
		cfg := base
		cfg.Maintainer = ""
		if err := ValidateDebConfig(cfg); err == nil {
			t.Fatal("expected error for missing maintainer, got nil")
		}
	})

	t.Run("missing description", func(t *testing.T) {
		cfg := base
		cfg.Description = ""
		if err := ValidateDebConfig(cfg); err == nil {
			t.Fatal("expected error for missing description, got nil")
		}
	})
}

func TestGenerateControlFile_HappyPath(t *testing.T) {
	cfg := DebConfig{
		Package:       "gowall",
		Version:       "v0.2.4",
		Architecture:  "amd64",
		Maintainer:    "Adam Magued <adamismailmageud@gmail.com>",
		Homepage:      "https://github.com/Achno/gowall",
		InstalledSize: 12500,
		Depends:       []string{"libc6 (>= 2.34)"},
		Description:   "Wallpaper color scheme and palette converter tool\nSecond line explaining details.\n\nFourth line after empty.",
	}

	ctrl, err := GenerateControlFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{
		"Package: gowall\n",
		"Version: 0.2.4\n",
		"Section: utils\n",
		"Priority: optional\n",
		"Architecture: amd64\n",
		"Installed-Size: 12500\n",
		"Maintainer: Adam Magued <adamismailmageud@gmail.com>\n",
		"Depends: libc6 (>= 2.34)\n",
		"Homepage: https://github.com/Achno/gowall\n",
		"Description: Wallpaper color scheme and palette converter tool\n",
		" Second line explaining details.\n",
		" .\n",
		" Fourth line after empty.\n",
	}

	for _, part := range expectedParts {
		if !strings.Contains(ctrl, part) {
			t.Errorf("control file missing expected fragment %q\nFull output:\n%s", part, ctrl)
		}
	}
}

func TestGenerateControlFile_ValidationFailure(t *testing.T) {
	cfg := DebConfig{}
	_, err := GenerateControlFile(cfg)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGeneratePackagesIndex_HappyPath(t *testing.T) {
	packages := []DebPackageInfo{
		{
			Package:       "gowall",
			Version:       "0.2.4",
			Architecture:  "amd64",
			Maintainer:    "Adam Magued <adamismailmageud@gmail.com>",
			InstalledSize: 15000,
			Filename:      "pool/main/g/gowall/gowall-amd64-linux.deb",
			Size:          5000000,
			SHA256:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Description:   "Wallpaper converter\nExtended description",
		},
		{
			Package:       "gowall",
			Version:       "0.2.4",
			Architecture:  "arm64",
			Maintainer:    "Adam Magued <adamismailmageud@gmail.com>",
			InstalledSize: 14500,
			Filename:      "pool/main/g/gowall/gowall-arm64-linux.deb",
			Size:          4800000,
			SHA256:        "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
			Description:   "Wallpaper converter arm64",
		},
	}

	index, err := GeneratePackagesIndex(packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(index, "Package: gowall\nVersion: 0.2.4\nArchitecture: amd64\n") {
		t.Errorf("index missing amd64 package block")
	}
	if !strings.Contains(index, "Package: gowall\nVersion: 0.2.4\nArchitecture: arm64\n") {
		t.Errorf("index missing arm64 package block")
	}
	if !strings.Contains(index, "Filename: pool/main/g/gowall/gowall-amd64-linux.deb\n") {
		t.Errorf("index missing filename")
	}
}

func TestGeneratePackagesIndex_Errors(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		_, err := GeneratePackagesIndex(nil)
		if err == nil {
			t.Fatal("expected error for empty packages list")
		}
	})

	t.Run("missing package name", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Version: "0.2.4", Architecture: "amd64", Filename: "file.deb", Size: 100, SHA256: "hash"},
		})
		if err == nil {
			t.Fatal("expected error for missing package name")
		}
	})

	t.Run("missing version", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Package: "gowall", Architecture: "amd64", Filename: "file.deb", Size: 100, SHA256: "hash"},
		})
		if err == nil {
			t.Fatal("expected error for missing version")
		}
	})

	t.Run("missing architecture", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Package: "gowall", Version: "0.2.4", Filename: "file.deb", Size: 100, SHA256: "hash"},
		})
		if err == nil {
			t.Fatal("expected error for missing architecture")
		}
	})

	t.Run("missing filename", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Package: "gowall", Version: "0.2.4", Architecture: "amd64", Size: 100, SHA256: "hash"},
		})
		if err == nil {
			t.Fatal("expected error for missing filename")
		}
	})

	t.Run("non-positive size", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Package: "gowall", Version: "0.2.4", Architecture: "amd64", Filename: "file.deb", Size: 0, SHA256: "hash"},
		})
		if err == nil {
			t.Fatal("expected error for zero size")
		}
	})

	t.Run("missing sha256", func(t *testing.T) {
		_, err := GeneratePackagesIndex([]DebPackageInfo{
			{Package: "gowall", Version: "0.2.4", Architecture: "amd64", Filename: "file.deb", Size: 100, SHA256: ""},
		})
		if err == nil {
			t.Fatal("expected error for missing sha256")
		}
	})
}

func TestGenerateReleaseIndex_HappyPath(t *testing.T) {
	cfg := RepoConfig{
		Origin:        "gowall",
		Label:         "gowall",
		Suite:         "stable",
		Codename:      "stable",
		Architectures: []string{"amd64", "arm64"},
		Components:    []string{"main"},
		Description:   "Gowall APT repository",
	}

	indices := []ComponentIndex{
		{
			Path:   "main/binary-amd64/Packages",
			Size:   1024,
			SHA256: "abc123def456",
		},
	}

	release, err := GenerateReleaseIndex(cfg, indices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Origin: gowall\n",
		"Label: gowall\n",
		"Suite: stable\n",
		"Codename: stable\n",
		"Architectures: amd64 arm64\n",
		"Components: main\n",
		"Description: Gowall APT repository\n",
		"SHA256:\n",
		" abc123def456 1024 main/binary-amd64/Packages\n",
	}

	for _, exp := range expected {
		if !strings.Contains(release, exp) {
			t.Errorf("Release missing fragment %q\nFull output:\n%s", exp, release)
		}
	}
}

func TestGenerateReleaseIndex_Errors(t *testing.T) {
	base := RepoConfig{
		Origin:        "gowall",
		Label:         "gowall",
		Suite:         "stable",
		Codename:      "stable",
		Architectures: []string{"amd64"},
		Components:    []string{"main"},
	}

	t.Run("missing origin", func(t *testing.T) {
		cfg := base
		cfg.Origin = ""
		if _, err := GenerateReleaseIndex(cfg, nil); err == nil {
			t.Fatal("expected error for missing origin")
		}
	})

	t.Run("missing architectures", func(t *testing.T) {
		cfg := base
		cfg.Architectures = nil
		if _, err := GenerateReleaseIndex(cfg, nil); err == nil {
			t.Fatal("expected error for missing architectures")
		}
	})

	t.Run("missing components", func(t *testing.T) {
		cfg := base
		cfg.Components = nil
		if _, err := GenerateReleaseIndex(cfg, nil); err == nil {
			t.Fatal("expected error for missing components")
		}
	})
}

func TestGenerateSourcesList_StandardAndSigned(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		cfg := RepoConfig{
			BaseURL:    "https://achno.github.io/gowall/apt",
			Suite:      "stable",
			Components: []string{"main"},
		}
		got, err := GenerateSourcesList(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "deb https://achno.github.io/gowall/apt stable main\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("signed", func(t *testing.T) {
		cfg := RepoConfig{
			BaseURL:       "https://achno.github.io/gowall/apt",
			Suite:         "stable",
			Components:    []string{"main"},
			SigningKeyURL: "/etc/apt/trusted.gpg.d/gowall.gpg",
		}
		got, err := GenerateSourcesList(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "deb [signed-by=/etc/apt/trusted.gpg.d/gowall.gpg] https://achno.github.io/gowall/apt stable main\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("missing base url", func(t *testing.T) {
		if _, err := GenerateSourcesList(RepoConfig{}); err == nil {
			t.Fatal("expected error for missing baseURL")
		}
	})
}

func TestGenerateDeb822Sources(t *testing.T) {
	cfg := RepoConfig{
		BaseURL:       "https://achno.github.io/gowall/apt",
		Suite:         "stable",
		Components:    []string{"main"},
		Architectures: []string{"amd64", "arm64"},
		SigningKeyURL: "/etc/apt/trusted.gpg.d/gowall.gpg",
	}

	got, err := GenerateDeb822Sources(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFragments := []string{
		"Types: deb\n",
		"URIs: https://achno.github.io/gowall/apt\n",
		"Suites: stable\n",
		"Components: main\n",
		"Signed-By: /etc/apt/trusted.gpg.d/gowall.gpg\n",
		"Architectures: amd64 arm64\n",
	}

	for _, frag := range expectedFragments {
		if !strings.Contains(got, frag) {
			t.Errorf("deb822 missing fragment %q\nFull output:\n%s", frag, got)
		}
	}
}

func TestComputeSHA256(t *testing.T) {
	input := []byte("hello gowall")
	expected := "50ef02959dc0789baab5c7bc0d5f7ec5d55c7e27e9d5d3a50125513c57633d6e"
	got := ComputeSHA256(input)
	if got != expected {
		t.Errorf("expected hash %q, got %q", expected, got)
	}
}
