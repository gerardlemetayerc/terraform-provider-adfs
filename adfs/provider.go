package adfs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/masterzen/winrm"
)

// Retourne le provider Terraform
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"adfs_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endpoint ADFS pour l'accès et l'authentification",
			},
			"winrm_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5986, // Port par défaut pour WinRM sur HTTPS
				Description: "Port utilisé pour se connecter via WinRM",
			},
			"winrm_username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nom d'utilisateur pour WinRM",
			},
			"winrm_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true, // Indique que le mot de passe est sensible
				Description: "Mot de passe pour WinRM",
			},
			"winrm_protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "HTTPS", // Protocole par défaut pour WinRM
				Description: "Protocole utilisé pour WinRM (HTTP/HTTPS)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"adfs_client":                   resourceAdfsClient(),
			"adfs_relying_party_trust":      resourceAdfsRelyingPartyTrust(),
			"adfs_relying_party_trust_rule": resourceAdfsRelyingPartyTrustRule(),
			"adfs_saml_endpoint":            resourceAdfsSamlEndpoint(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// Fonction de configuration du provider
func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	adfsEndpoint := d.Get("adfs_endpoint").(string)
	winrmPort := d.Get("winrm_port").(int)
	winrmUsername := d.Get("winrm_username").(string)
	winrmPassword := d.Get("winrm_password").(string)
	winrmProtocol := d.Get("winrm_protocol").(string)

	endpoint := fmt.Sprintf("%s:%d", adfsEndpoint, winrmPort)

	client, err := winrm.NewClient(&winrm.Endpoint{
		Host:     endpoint,
		HTTPS:    (winrmProtocol == "HTTPS"),
		Insecure: false,
		Port:     winrmPort,
	}, winrmUsername, winrmPassword)

	if err != nil {
		return nil, err
	}

	config := &AdfsConfig{
		Client: client, // Client WinRM configuré
	}

	return config, nil
}

// Structure de configuration pour le provider
type AdfsConfig struct {
	Client *winrm.Client // Client WinRM
}
