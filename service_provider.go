package vite

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/path"
)

const Binding = "goravel.vite"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	vite := NewVite(app.MakeConfig())

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return vite, nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {

	route := app.MakeRoute()
	config := app.MakeConfig()
	route.Static(config.GetString("vite.base_url", "/static"), path.Base(config.GetString("vite.assets_path", "public/build")))

	app.Publishes("github.com/merouanekhalili/goravel-vite", map[string]string{
		"config/vite.go":                       app.ConfigPath("vite.go"),
		"templates/.prettierignore.txt":        path.Base(".prettierignore"),
		"templates/.prettierrc.txt":            path.Base(".prettierrc"),
		"templates/react/views":                path.Base("resources/views"),
		"templates/react/js/App.tsx.txt":       path.Base("resources/js/App.tsx"),
		"templates/react/js/main.tsx.txt":      path.Base("resources/js/main.tsx"),
		"templates/react/css/app.css.txt":      path.Base("resources/css/app.css"),
		"templates/react/vite.config.ts.txt":   path.Base("vite.config.ts"),
		"templates/react/package.json.txt":     path.Base("package.json"),
		"templates/react/tsconfig.json.txt":    path.Base("tsconfig.json"),
		"templates/react/components.json.txt":  path.Base("components.json"),
		"templates/react/eslint.config.js.txt": path.Base("eslint.config.js"),
	}, "react")

	app.Publishes("github.com/merouanekhalili/goravel-vite", map[string]string{
		"config/vite.go":                     app.ConfigPath("vite.go"),
		"templates/.prettierignore.txt":      path.Base(".prettierignore"),
		"templates/.prettierrc.txt":          path.Base(".prettierrc"),
		"templates/vue/views":                path.Base("resources/views"),
		"templates/vue/js/App.vue.txt":       path.Base("resources/js/App.vue"),
		"templates/vue/js/main.ts.txt":       path.Base("resources/js/main.ts"),
		"templates/vue/js/env.d.ts.txt":      path.Base("resources/js/env.d.ts"),
		"templates/vue/css/app.css.txt":      path.Base("resources/css/app.css"),
		"templates/vue/vite.config.ts.txt":   path.Base("vite.config.ts"),
		"templates/vue/package.json.txt":     path.Base("package.json"),
		"templates/vue/tsconfig.json.txt":    path.Base("tsconfig.json"),
		"templates/vue/components.json.txt":  path.Base("components.json"),
		"templates/vue/eslint.config.js.txt": path.Base("eslint.config.js"),
	}, "vue")
}
