package facades

import (
	vite "github.com/merouanekhalili/goravel-vite"
	"github.com/merouanekhalili/goravel-vite/contracts"
)

func Vite() (contracts.Vite, error) {
	instance, err := vite.App.Make(vite.Binding)
	if err != nil {
		return nil, err
	}

	return instance.(contracts.Vite), nil
}
