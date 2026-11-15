package providers

import (
	"strings"
	"testing"
)

func TestResolvePrompt(t *testing.T) {
	raw := ResolvePrompt(ProviderConfig{
		Prompt: "Text Recognition:", AppendPrompt: "more", Format: "md", RawPrompt: true,
	}, "scan.pdf-page-2-of-5")
	if raw != "Text Recognition:" {
		t.Fatalf("raw prompt = %q", raw)
	}

	built := ResolvePrompt(ProviderConfig{
		Prompt: "Text Recognition:", AppendPrompt: "more", Format: "md",
	}, "scan.pdf-page-2-of-5")
	for _, want := range []string{"Text Recognition:", "Format the output in Markdown.", "PAGE 2", "more"} {
		if !strings.Contains(built, want) {
			t.Fatalf("built prompt %q missing %q", built, want)
		}
	}
}
