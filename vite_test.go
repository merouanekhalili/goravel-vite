package vite

import (
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

type ViteTestSuite struct {
	suite.Suite

	mockConfig *mocksconfig.Config
	vite       *Vite
	tempDir    string
}

func resetGlobals() {
	manifestOnce = sync.Once{}
	entryPointsOnce = sync.Once{}
	manifest = nil
	manifestErr = nil
	entryPoints = nil
}

func (s *ViteTestSuite) SetupTest() {
	resetGlobals()

	s.mockConfig = mocksconfig.NewConfig(s.T())

	s.vite = NewVite(s.mockConfig)

	dir, err := os.MkdirTemp("", "vite_test_manifest_")
	s.Require().NoError(err)
	s.tempDir = dir

}

func (s *ViteTestSuite) TearDownTest() {

	os.RemoveAll(s.tempDir)
}

func (s *ViteTestSuite) writeManifest(content string) {

	manifestPath := s.mockConfig.GetString("vite.manifest_path", "public/build/.vite/manifest.json")
	err := os.WriteFile(manifestPath, []byte(content), 0644)
	s.Require().NoError(err, "Failed to write mock manifest file")
}

func TestViteTestSuite(t *testing.T) {
	suite.Run(t, new(ViteTestSuite))
}

func (s *ViteTestSuite) TestAssets_LocalEnvironment_DefaultFramework() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("local").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.dev_server_url", "http://localhost:5173").Return("http://localhost:5173").Once()
	s.mockConfig.On("GetString", "vite.entry_points", "").Return("resources/js/app.js").Once()

	expected := template.HTML(`<script type="module" src="http://localhost:5173/@vite/client"></script><script type="module" src="http://localhost:5173/resources/js/app.js"></script>`)
	actual := s.vite.Assets()

	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_LocalEnvironment_ReactFramework() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("local").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("react").Once()
	s.mockConfig.On("GetString", "vite.dev_server_url", "http://localhost:5173").Return("http://localhost:5173").Once()
	s.mockConfig.On("GetString", "vite.entry_points", "").Return("resources/js/app.jsx").Once()

	actual := s.vite.Assets()
	htmlString := string(actual)

	assert.Contains(s.T(), htmlString, `import RefreshRuntime from "http://localhost:5173/@react-refresh";`)
	assert.Contains(s.T(), htmlString, `window.__vite_plugin_react_preamble_installed__ = true;`)

	assert.Contains(s.T(), htmlString, `<script type="module" src="http://localhost:5173/@vite/client"></script>`)

	assert.Contains(s.T(), htmlString, `<script type="module" src="http://localhost:5173/resources/js/app.jsx"></script>`)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_SingleEntryPoint_NoCSS() {
	entryPoint := "resources/js/app.js"
	manifestContent := `{
		"resources/js/app.js": {
			"file": "assets/app.12345.js",
			"src": "resources/js/app.js",
			"isEntry": true
		}
	}`

	s.mockConfig.ExpectedCalls = nil

	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()

	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()

	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)

	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPoint).Once()

	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return("/static/").Maybe()

	expected := template.HTML(`<link rel="modulepreload" href="/static/assets/app.12345.js"><script type="module" src="/static/assets/app.12345.js"></script>`)
	actual := s.vite.Assets()

	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_SingleEntryPoint_WithCSS() {
	entryPoint := "resources/js/app.js"
	manifestContent := `{
		"resources/js/app.js": {
			"file": "assets/app.12345.js",
			"src": "resources/js/app.js",
			"isEntry": true,
			"css": ["assets/app.67890.css"]
		}
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPoint).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return("/static/").Maybe()

	expected := template.HTML(`<link rel="modulepreload" href="/static/assets/app.12345.js"><link rel="preload" href="/static/assets/app.67890.css" as="style"><script type="module" src="/static/assets/app.12345.js"></script><link rel="stylesheet" href="/static/assets/app.67890.css">`)
	actual := s.vite.Assets()

	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_MultipleEntryPoints() {
	entryPointsStr := "resources/js/app.js,resources/js/admin.js"
	manifestContent := `{
		"resources/js/app.js": {
			"file": "assets/app.12345.js",
			"src": "resources/js/app.js",
			"isEntry": true,
			"css": ["assets/app.abcde.css"]
		},
		"resources/js/admin.js": {
			"file": "assets/admin.67890.js",
			"src": "resources/js/admin.js",
			"isEntry": true,
			"css": ["assets/admin.fghij.css"]
		}
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPointsStr).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return("/static/").Maybe()

	actual := s.vite.Assets()
	htmlString := string(actual)

	assert.Contains(s.T(), htmlString, `<link rel="modulepreload" href="/static/assets/app.12345.js">`)
	assert.Contains(s.T(), htmlString, `<link rel="modulepreload" href="/static/assets/admin.67890.js">`)
	assert.Contains(s.T(), htmlString, `<link rel="preload" href="/static/assets/app.abcde.css" as="style">`)
	assert.Contains(s.T(), htmlString, `<link rel="preload" href="/static/assets/admin.fghij.css" as="style">`)

	assert.Contains(s.T(), htmlString, `<script type="module" src="/static/assets/app.12345.js"></script>`)
	assert.Contains(s.T(), htmlString, `<link rel="stylesheet" href="/static/assets/app.abcde.css">`)
	assert.Contains(s.T(), htmlString, `<script type="module" src="/static/assets/admin.67890.js"></script>`)
	assert.Contains(s.T(), htmlString, `<link rel="stylesheet" href="/static/assets/admin.fghij.css">`)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_WithImports() {
	entryPoint := "resources/js/app.js"
	manifestContent := `{
		"resources/js/app.js": {
			"file": "assets/app.12345.js",
			"src": "resources/js/app.js",
			"isEntry": true,
			"imports": ["_vendor.abcdef.js"]
		},
		"_vendor.abcdef.js": {
			"file": "assets/vendor.abcdef.js"
		}
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPoint).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return("/static/").Maybe()

	expected := template.HTML(`<link rel="modulepreload" href="/static/assets/app.12345.js"><link rel="modulepreload" href="/static/assets/vendor.abcdef.js"><script type="module" src="/static/assets/app.12345.js"></script>`)
	actual := s.vite.Assets()

	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_ManifestNotFound() {

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.entry_points", "").Return("resources/js/app.js").Once()

	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Once()

	actual := s.vite.Assets()
	htmlString := string(actual)

	assert.Contains(s.T(), htmlString, "<!-- ERROR: Could not load Vite manifest:")
	assert.Contains(s.T(), htmlString, "reading manifest file")
}

func (s *ViteTestSuite) TestAssets_Production_InvalidManifest() {
	manifestContent := `{"invalid json`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.entry_points", "").Return("resources/js/app.js").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)

	actual := s.vite.Assets()
	htmlString := string(actual)

	assert.Contains(s.T(), htmlString, "<!-- ERROR: Could not load Vite manifest:")
	assert.Contains(s.T(), htmlString, "parsing manifest JSON")
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_Production_EntryPointNotInManifest() {
	missingEntryPoint := "resources/js/missing.js"
	manifestContent := `{
		"resources/js/app.js": {
			"file": "assets/app.12345.js",
			"src": "resources/js/app.js",
			"isEntry": true
		}
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(missingEntryPoint).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return("/static/").Maybe()

	expected := template.HTML(``)
	actual := s.vite.Assets()

	assert.Equal(s.T(), expected, actual, "Should render empty string if entry point is missing")
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_BaseURL_TrailingSlash() {
	entryPoint := "resources/js/app.js"
	baseURL := "/custom/static/"
	manifestContent := `{
		"resources/js/app.js": { "file": "assets/app.12345.js", "src": "resources/js/app.js", "isEntry": true }
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPoint).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return(baseURL).Maybe()

	expected := template.HTML(`<link rel="modulepreload" href="/custom/static/assets/app.12345.js"><script type="module" src="/custom/static/assets/app.12345.js"></script>`)
	actual := s.vite.Assets()
	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ViteTestSuite) TestAssets_BaseURL_NoTrailingSlash() {
	entryPoint := "resources/js/app.js"
	baseURL := "/custom/static"
	manifestContent := `{
		"resources/js/app.js": { "file": "assets/app.12345.js", "src": "resources/js/app.js", "isEntry": true }
	}`

	s.mockConfig.ExpectedCalls = nil
	s.mockConfig.On("GetString", "app.env", "production").Return("production").Once()
	s.mockConfig.On("GetString", "vite.js_framework", "vue").Return("vue").Once()
	s.mockConfig.On("GetString", "vite.manifest_path", "public/build/.vite/manifest.json").Return(filepath.Join(s.tempDir, "manifest.json")).Twice()
	s.writeManifest(manifestContent)
	s.mockConfig.On("GetString", "vite.entry_points", "").Return(entryPoint).Once()
	s.mockConfig.On("GetString", "vite.base_url", "/static/").Return(baseURL).Maybe()

	expected := template.HTML(`<link rel="modulepreload" href="/custom/static/assets/app.12345.js"><script type="module" src="/custom/static/assets/app.12345.js"></script>`)
	actual := s.vite.Assets()
	assert.Equal(s.T(), expected, actual)
	s.mockConfig.AssertExpectations(s.T())
}
