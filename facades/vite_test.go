package facades

import (
	"testing"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vite "github.com/merouanekhalili/goravel-vite"
)

func TestMakingSessionWithRealImplementation(t *testing.T) {

	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)

	originalApp := vite.App
	defer func() {
		vite.App = originalApp
	}()
	vite.App = mockApp

	realViteInstance := vite.NewVite(mockConfig)

	mockApp.EXPECT().Make(vite.Binding).
		Return(realViteInstance, nil).Once()

	instance, err := Vite()
	require.NoError(t, err)

	require.NotNil(t, instance)
	assert.Equal(t, realViteInstance, instance)
}
