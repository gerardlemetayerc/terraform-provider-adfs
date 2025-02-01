package adfs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Retourne le provider Terraform
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"powershell_bin": {
				Type:        schema.TypeString,
				Required:    true,
				Default:     "powershell",
				Description: "Powershell binary",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"adfs_relying_party_trust": resourceAdfsRelyingPartyTrust(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// Fonction de configuration du provider
func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	powershell_bin := d.Get("powershell_bin").(string)

	config := &AdfsConfig{
		PowershellBin: powershell_bin,
	}

	return config, nil
}

// Structure de configuration pour le provider
type AdfsConfig struct {
	PowershellBin string
}
