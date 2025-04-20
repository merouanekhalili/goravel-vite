package contracts

import "html/template"

type Vite interface {
	Assets() template.HTML
}
