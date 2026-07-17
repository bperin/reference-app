package auth

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSwaggerOAuth2AuthContract(t *testing.T) {
	data, err := os.ReadFile("../../docs/swagger.json")
	if err != nil {
		t.Fatalf("read generated Swagger spec: %v", err)
	}

	var spec struct {
		SecurityDefinitions map[string]struct {
			Type     string `json:"type"`
			Flow     string `json:"flow"`
			TokenURL string `json:"tokenUrl"`
		} `json:"securityDefinitions"`
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("decode generated Swagger spec: %v", err)
	}

	oauth, ok := spec.SecurityDefinitions["OAuth2Auth"]
	if !ok {
		t.Fatal("OAuth2Auth security definition is missing")
	}
	if oauth.Type != "oauth2" || oauth.Flow != "password" || oauth.TokenURL != "/auth/oauth/token" {
		t.Fatalf("OAuth2Auth = %#v, want password flow at /auth/oauth/token", oauth)
	}
	if _, ok := spec.Paths["/auth/oauth/token"]["post"]; !ok {
		t.Fatal("POST /auth/oauth/token is missing from generated Swagger spec")
	}
	if _, ok := spec.Paths["/auth/register"]["post"]; !ok {
		t.Fatal("POST /auth/register is missing from generated Swagger spec")
	}

	for _, route := range []struct {
		path   string
		method string
	}{
		{path: "/users/me", method: "get"},
		{path: "/posts/", method: "post"},
	} {
		operation, ok := spec.Paths[route.path][route.method]
		if !ok {
			t.Fatalf("%s %s is missing from generated Swagger spec", route.method, route.path)
		}
		var documented struct {
			Security []map[string][]string `json:"security"`
		}
		if err := json.Unmarshal(operation, &documented); err != nil {
			t.Fatalf("decode %s %s operation: %v", route.method, route.path, err)
		}
		if len(documented.Security) != 1 || len(documented.Security[0]["OAuth2Auth"]) != 1 || documented.Security[0]["OAuth2Auth"][0] != "user" {
			t.Fatalf("%s %s security = %#v, want OAuth2Auth[user]", route.method, route.path, documented.Security)
		}
	}
}
