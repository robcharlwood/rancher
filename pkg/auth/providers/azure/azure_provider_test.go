package azure

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rancher/norman/api/writer"
	"github.com/rancher/norman/types"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementschema "github.com/rancher/rancher/pkg/schemas/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
)

// TestConfigureTest inspects the Redirect URL during Azure AD setup.
func TestConfigureTest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		authConfig          map[string]interface{}
		expectedRedirectURL string
	}{
		{
			name: "initial setup of Azure AD with Microsoft Graph",
			authConfig: map[string]interface{}{
				"accessMode": "unrestricted",
				"annotations": map[string]interface{}{
					"auth.cattle.io/azuread-endpoint-migrated": "true",
				},
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.microsoft.com",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/v2.0/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectURL: "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
		{
			name: "attempt to initially setup Azure AD with deprecated Azure AD Graph",
			authConfig: map[string]interface{}{
				"accessMode":        "unrestricted",
				"annotations":       map[string]interface{}{},
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.windows.net/",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectURL: "https://login.microsoftonline.com/tenant123/oauth2/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
		{
			name: "editing an existing setup of Azure AD",
			authConfig: map[string]interface{}{
				"enabled":    true,
				"accessMode": "unrestricted",
				"annotations": map[string]interface{}{
					"auth.cattle.io/azuread-endpoint-migrated": "true",
				},
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.microsoft.com",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/v2.0/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectURL: "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
		{
			name: "editing an existing setup of Azure AD without annotation",
			authConfig: map[string]interface{}{
				"enabled":           true,
				"accessMode":        "unrestricted",
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.windows.net/",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectURL: "https://login.microsoftonline.com/tenant123/oauth2/authorize?client_id=app123&redirect_uri=https://myrancher.com&resource=https://graph.windows.net/&scope=openid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := json.Marshal(test.authConfig)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v3/azureADConfigs/azuread?action=configureTest", bytes.NewReader(b))

			schemas := types.NewSchemas()
			schemas.AddSchemas(managementschema.AuthSchemas)

			rw := &writer.EncodingResponseWriter{
				ContentType: "application/json",
				Encoder:     types.JSONEncoder,
			}
			rr := httptest.NewRecorder()
			r := &types.APIContext{
				Schemas:        schemas,
				Request:        req,
				Response:       rr,
				ResponseWriter: rw,
				Version:        &managementschema.Version,
			}

			provider := azureProvider{}
			err = provider.ConfigureTest("configureTest", nil, r)
			assert.NoError(t, err)

			res := rr.Result()
			defer res.Body.Close()

			var output v3.AzureADConfigTestOutput
			err = json.NewDecoder(res.Body).Decode(&output)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedRedirectURL, output.RedirectURL)
		})
	}

}

func TestTransformToAuthProvider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		authConfig          map[string]interface{}
		expectedRedirectUrl string
	}{
		{
			name: "redirect URL for Microsoft Graph",
			authConfig: map[string]interface{}{
				"enabled":    true,
				"accessMode": "unrestricted",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"auth.cattle.io/azuread-endpoint-migrated": "true",
					},
				},
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.microsoft.com",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/v2.0/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectUrl: "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
		{
			name: "redirect URL for Azure AD Graph",
			authConfig: map[string]interface{}{
				"enabled":           true,
				"accessMode":        "unrestricted",
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.windows.net/",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectUrl: "https://login.microsoftonline.com/tenant123/oauth2/authorize?client_id=app123&redirect_uri=https://myrancher.com&resource=https://graph.windows.net/&scope=openid",
		},
		{
			name: "redirect URL for disabled auth provider with annotation",
			authConfig: map[string]interface{}{
				"accessMode": "unrestricted",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"auth.cattle.io/azuread-endpoint-migrated": "true",
					},
				},
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.microsoft.com",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectUrl: "https://login.microsoftonline.com/tenant123/oauth2/v2.0/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
		{
			name: "redirect URL for disabled auth provider without annotation",
			authConfig: map[string]interface{}{
				"enabled":           false, // Here, enabled is set to false explicitly.
				"accessMode":        "unrestricted",
				"endpoint":          "https://login.microsoftonline.com/",
				"graphEndpoint":     "https://graph.windows.net/",
				"tokenEndpoint":     "https://login.microsoftonline.com/tenant123/oauth2/token",
				"authEndpoint":      "https://login.microsoftonline.com/tenant123/oauth2/authorize",
				"tenantId":          "tenant123",
				"applicationId":     "app123",
				"applicationSecret": "secret123",
				"rancherUrl":        "https://myrancher.com",
			},
			expectedRedirectUrl: "https://login.microsoftonline.com/tenant123/oauth2/authorize?client_id=app123&redirect_uri=https://myrancher.com&response_type=code&scope=openid",
		},
	}

	var provider azureProvider
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authProvider, err := provider.TransformToAuthProvider(test.authConfig)
			assert.NoError(t, err)
			url, ok := authProvider["redirectUrl"].(string)
			assert.True(t, ok)
			assert.Equal(t, test.expectedRedirectUrl, url)
		})
	}
}
