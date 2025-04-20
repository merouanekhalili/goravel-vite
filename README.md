# Goravel Vite Integration

Integrate Vite seamlessly into your Goravel application for modern frontend asset management. This package provides helpers to load Vite-processed assets (including CSS and JS for frameworks like React or Vue) in both development and production environments.

## Features

- Automatic loading of assets from Vite Dev Server in development.
- Automatic loading of versioned/hashed assets from the manifest file in production.
- Support for React (including Fast Refresh) and Vue.
- Publishable configuration and frontend scaffolding.
- Configurable via environment variables.

## Installation

1.  **Require the package:**

    ```bash
    go get github.com/merouanekhalili/goravel-vite
    ```

2.  **Register the Service Provider:**
    Add the provider to your `config/app.go` file within the `Providers` array:

    ```go
    // config/app.go
    package config

    import (
        "goravel/app/providers"
        // ... other imports
        vite "github.com/merouanekhalili/goravel-vite" // Add this import
    )

    func init() {
        // ...
        config.Add("app", map[string]any{
            // ...
            "providers": []contractsfoundation.ServiceProvider{
                // ... other providers
                &vite.ServiceProvider{}, // Add this line
            },
        })
    }
    ```

## Setup

1.  **Publish Assets:**
    Publish the configuration file and frontend scaffolding using the Artisan command. Choose the tag corresponding to your desired frontend framework (`react` or `vue`):

    ```bash
    # For React
    go run . artisan vendor:publish --package=github.com/merouanekhalili/goravel-vite --tag=react

    # For Vue
    go run . artisan vendor:publish --package=github.com/merouanekhalili/goravel-vite --tag=vue
    ```

    This command will:

    - Create `config/vite.go`.
    - Create necessary frontend files (`package.json`, `vite.config.ts`, `resources/js/...`, `resources/css/...`, etc.).
    - Create `.prettierrc`, `.prettierignore`.

2.  **Install Frontend Dependencies:**
    Navigate to your project root and install the Node dependencies:

    ```bash
    npm install
    # or
    yarn install
    ```

3.  **Configure Environment Variables:**
    Update your `.env` file. Only `VITE_JS_FRAMEWORK` and `VITE_ENTRY_POINTS` are typically needed if you keep the defaults for other settings:

    ```dotenv
    # .env
    VITE_JS_FRAMEWORK=vue
    VITE_ENTRY_POINTS=resources/js/main.ts

    # Other variables (defaults shown, uncomment/adjust if needed):
    # APP_ENV=local # 'local' for development, 'production' for production
    # VITE_DEV_SERVER_URL=http://localhost:5173
    # VITE_ASSETS_PATH=public/build
    # VITE_MANIFEST_PATH=public/build/.vite/manifest.json
    # VITE_BASE_URL=/static
    ```

## Usage

1.  **Share Vite Assets Globally (Recommended):**
    To make Vite assets available to all your views easily, share them globally from your `app/providers/app_service_provider.go`. This is the recommended approach, especially when using the template provided by this package.

    ```go
    // app/providers/app_service_provider.go
    package providers

    import (
    	"github.com/goravel/framework/contracts/foundation"
    	"github.com/goravel/framework/facades"
    	vitefacades "github.com/merouanekhalili/goravel-vite/facades" // Import Vite facade
    )

    type AppServiceProvider struct {
    }

    func (receiver *AppServiceProvider) Register(app foundation.Application) {
    	// ...
    }

    func (receiver *AppServiceProvider) Boot(app foundation.Application) {
        // Share Vite assets with all views using the key "vite"
        facades.View().Share("vite", vitefacades.Vite().Assets())

    	// Register other boot logic
    }

    ```

2.  **Use the Provided Template (`resources/views/app.tmpl`):**
    The `vendor:publish` command (from the Setup section) creates a template file at `resources/views/app.tmpl`. This template is designed to work with the global sharing method described above, as it already includes `{{ .vite }}` to load the assets.

    Here's a simplified view of the relevant part of `resources/views/app.tmpl`:

    ```html
    {{ define "app.tmpl" }}
    <!DOCTYPE html>
    <html lang="en">
      <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Goravel App</title>

        <!-- Vite assets are loaded here via the globally shared variable -->
        {{ .vite }}
      </head>
      <body>
        <!-- Your frontend app attaches here (e.g., #app or #app-root) -->
        <div id="app"></div>
        {{/* Or
        <div id="app-root"></div>
        for React */}}
      </body>
    </html>
    {{ end }}
    ```

    Simply ensure your controller/route returns this template. For example, in your `routes/web.go`:

    ```go
    // routes/web.go
    package routes

    import (
    	"github.com/goravel/framework/contracts/http"
    	"github.com/goravel/framework/facades"
    )

    func Web() {
    	facades.Route().Get("/", func(ctx http.Context) http.Response {
    		// Return the app.tmpl view
    		// The "vite" variable is automatically available thanks to AppServiceProvider
    		return ctx.Response().View().Make("app.tmpl", map[string]any{
    			// You can pass additional data to your view here
    			"name": "Goravel",
    		})
    	})

    	// Define other routes
    }
    ```

    The `{{ .vite }}` tag will automatically receive the shared assets and output the correct `<script>` and `<link>` tags based on your `APP_ENV`.

3.  **Run Development Servers:**
    Start the Vite development server and the Goravel development server in separate terminals:

    ```bash
    # Terminal 1: Start Vite Dev Server
    npm run dev
    ```

    ```bash
    # Terminal 2: Start Goravel Server
    go run .
    ```

    Now you can access your Goravel application, and assets will be loaded via the Vite dev server with Hot Module Replacement (HMR).

4.  **Build for Production:**
    When deploying, first build your frontend assets using Vite:

    ```bash
    npm run build
    ```

    This will generate optimized assets and a `manifest.json` file in the directory specified by `VITE_ASSETS_PATH` (default: `public/build`).

    Then, ensure your Goravel application's `APP_ENV` is set to `production`. The Vite helper (whether called directly or via the shared variable) will now use the `manifest.json` to load the correct, hashed asset files and serve them via the static route configured by the service provider (default prefix `/static`).

## Configuration Reference (`config/vite.go`)

- `js_framework`: (`VITE_JS_FRAMEWORK`, default: `"vue"`) - Sets the JS framework ("vue" or "react"). Determines scaffolding and React HMR setup.
- `entry_points`: (`VITE_ENTRY_POINTS`, default: `"resources/js/main.tsx"`) - Comma-separated list of main entry files for Vite.
- `dev_server_url`: (`VITE_DEV_SERVER_URL`, default: `"http://localhost:5173"`) - URL of the Vite dev server.
- `assets_path`: (`VITE_ASSETS_PATH`, default: `"public/build"`) - Directory where Vite places built assets. Must be publicly accessible.
- `manifest_path`: (`VITE_MANIFEST_PATH`, default: `"public/build/.vite/manifest.json"`) - Path to the generated Vite manifest file.
- `base_url`: (`VITE_BASE_URL`, default: `"/static"`) - Base URL prefix for serving built assets in production.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
