package webui

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestTemplatesAreCSPCompatible guards against inline scripts and inline event
// handlers, which the Content-Security-Policy (script-src 'self') blocks.
func TestTemplatesAreCSPCompatible(t *testing.T) {
	entries, err := fs.ReadDir(templateFiles, "templates")
	if err != nil {
		t.Fatalf("read templates: %v", err)
	}

	scriptTag := regexp.MustCompile(`(?i)<script\b[^>]*>`)
	inlineHandler := regexp.MustCompile(`(?i)\son[a-z]+\s*=\s*["']`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := fs.ReadFile(templateFiles, "templates/"+entry.Name())
		if err != nil {
			t.Fatalf("read template %s: %v", entry.Name(), err)
		}
		text := string(content)

		for _, tag := range scriptTag.FindAllString(text, -1) {
			if !strings.Contains(strings.ToLower(tag), "src=") {
				t.Errorf("%s: inline <script> is blocked by CSP; move it to /static/js", entry.Name())
			}
		}
		if inlineHandler.MatchString(text) {
			t.Errorf("%s: inline event handler is blocked by CSP; use data attributes and an external script", entry.Name())
		}
		if strings.Contains(text, `target="_blank"`) {
			t.Errorf("%s: target=\"_blank\" is not allowed; keep navigation in the same tab", entry.Name())
		}
	}
}

// TestStaticAssetsIncludeScripts verifies that the external scripts referenced
// by the templates are embedded and served under /static.
func TestStaticAssetsIncludeScripts(t *testing.T) {
	static, err := StaticFS()
	if err != nil {
		t.Fatalf("static filesystem: %v", err)
	}
	for _, name := range []string{
		"css/style.css",
		"js/photo-viewer.js",
		"js/publish-preview.js",
		"js/contact-draft.js",
		"js/confirm-actions.js",
	} {
		if _, err := fs.Stat(static, name); err != nil {
			t.Errorf("missing static asset %s: %v", name, err)
		}
	}
}
