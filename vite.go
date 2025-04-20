package vite

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"sync"

	"github.com/merouanekhalili/goravel-vite/contracts"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/support/path"
)

type viteManifest map[string]viteManifestEntry
type viteManifestEntry struct {
	File    string   `json:"file"`
	IsEntry bool     `json:"isEntry,omitempty"`
	Src     string   `json:"src,omitempty"`
	CSS     []string `json:"css,omitempty"`
	Assets  []string `json:"assets,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

var (
	manifest        viteManifest
	manifestOnce    sync.Once
	manifestErr     error
	entryPoints     []string
	entryPointsOnce sync.Once
)

var _ contracts.Vite = &Vite{}

type Vite struct {
	config config.Config
}

func NewVite(config config.Config) *Vite {
	return &Vite{config: config}
}

func (v *Vite) Assets() template.HTML {

	env := v.config.GetString("app.env", "production")
	jsFramework := v.config.GetString("vite.js_framework", "vue")

	entryPointsOnce.Do(func() {
		entryPoints = strings.Split(v.config.GetString("vite.entry_points", ""), ",")
	})

	var sb strings.Builder

	if env == "local" {

		viteDevServer := v.config.GetString("vite.dev_server_url", "http://localhost:5173")

		if jsFramework == "react" {
			sb.WriteString(`<script type="module">
			import RefreshRuntime from "` + viteDevServer + `/@react-refresh";
			RefreshRuntime.injectIntoGlobalHook(window);
			window.$RefreshReg$ = () => {};
			window.$RefreshSig$ = () => (type) => type;
			window.__vite_plugin_react_preamble_installed__ = true;
			</script>`)
		}

		sb.WriteString(fmt.Sprintf(`<script type="module" src="%s/@vite/client"></script>`, viteDevServer))

		for _, entry := range entryPoints {
			sb.WriteString(fmt.Sprintf(`<script type="module" src="%s/%s"></script>`, viteDevServer, entry))
		}

	} else {

		manifest, manifestErr = v.loadManifest()
		if manifestErr != nil {
			return template.HTML(fmt.Sprintf("<!-- ERROR: Could not load Vite manifest: %v -->", manifestErr))
		}

		includedCSS := make(map[string]bool)
		includedJSPreload := make(map[string]bool)
		includedCSSPreload := make(map[string]bool)
		baseURL := v.config.GetString("vite.base_url", "/static/")

		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}

		var preloadJS func(string)
		preloadJS = func(moduleSrc string) {
			if includedJSPreload[moduleSrc] {
				return
			}
			entry, ok := manifest[moduleSrc]
			if !ok {
				return
			}

			jsPath := baseURL + entry.File
			sb.WriteString(fmt.Sprintf(`<link rel="modulepreload" href="%s">`, jsPath))
			includedJSPreload[moduleSrc] = true

			for _, imp := range entry.Imports {
				preloadJS(imp)
			}
		}

		for _, entrySrc := range entryPoints {
			entry, ok := manifest[entrySrc]
			if !ok {
				continue
			}

			preloadJS(entrySrc)

			for _, cssFile := range entry.CSS {
				if !includedCSSPreload[cssFile] {
					cssPath := baseURL + cssFile
					sb.WriteString(fmt.Sprintf(`<link rel="preload" href="%s" as="style">`, cssPath))
					includedCSSPreload[cssFile] = true
				}
			}
		}

		for _, entrySrc := range entryPoints {
			entry, ok := manifest[entrySrc]
			if !ok {
				continue
			}

			if strings.HasSuffix(strings.ToLower(entry.File), ".js") {
				jsPath := baseURL + entry.File
				sb.WriteString(fmt.Sprintf(`<script type="module" src="%s"></script>`, jsPath))
			}

			for _, cssFile := range entry.CSS {
				if !includedCSS[cssFile] {
					cssPath := baseURL + cssFile
					sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssPath))
					includedCSS[cssFile] = true
				}
			}
		}
	}

	return template.HTML(sb.String())
}

func (v *Vite) loadManifest() (viteManifest, error) {
	manifestOnce.Do(func() {

		manifestPath := path.Base(v.config.GetString("vite.manifest_path", "public/build/.vite/manifest.json"))

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			manifestErr = fmt.Errorf("reading manifest file %q: %w", manifestPath, err)
			return
		}

		var m viteManifest
		err = json.Unmarshal(data, &m)
		if err != nil {
			manifestErr = fmt.Errorf("parsing manifest JSON %q: %w", manifestPath, err)
			return
		}
		manifest = m
		manifestErr = nil
	})
	return manifest, manifestErr
}
