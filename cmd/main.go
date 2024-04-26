package main

import (
	"github.com/gerardlemetayerc/terraform-provider-adfs/adfs" // Pour gérer les schémas Terraform
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"      // Pour démarrer le provider
)

// Point d'entrée pour le programme Go
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: adfs.Provider,
	})
}
