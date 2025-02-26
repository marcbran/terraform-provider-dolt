package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dolt": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
}

func testAccProviderConfig(path string) string {
	return fmt.Sprintf(`
provider "dolt" {
  path  = %[1]q
  email = "test@example.com"
  name  = "Test Example"
}
`, path)
}
