package adfs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAdfsSamlEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdfsSamlEndpointCreate,
		Read:   resourceAdfsSamlEndpointRead,
		Update: resourceAdfsSamlEndpointUpdate,
		Delete: resourceAdfsSamlEndpointDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"relying_party_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"binding": {
				Type:     schema.TypeString,
				Required: true,
			},
			"index": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
		},
	}
}

func resourceAdfsSamlEndpointCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	relyingPartyIdentifier := d.Get("relying_party_identifier").(string)
	endpointUrl := d.Get("endpoint_url").(string)
	binding := d.Get("binding").(string)
	index := d.Get("index").(int)

	command := fmt.Sprintf(
		"Add-AdfsRelyingPartyEndpoint -Name '%s' -RelyingPartyIdentifier '%s' -Endpoint '%s' -Binding '%s' -Index %d",
		name, relyingPartyIdentifier, endpointUrl, binding, index,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la création du SAML endpoint : %v", err)
	}

	d.SetId(name)

	return nil
}

func resourceAdfsSamlEndpointRead(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"$endpoint = Get-AdfsRelyingPartyEndpoint -Name '%s' "+
			"if ($endpoint) { "+
			"return @{ 'name' = $endpoint.Name; 'relying_party_identifier' = $endpoint.RelyingPartyIdentifier; "+
			"'endpoint_url' = $endpoint.Endpoint; 'binding' = $endpoint.Binding; 'index' = $endpoint.Index } "+
			"} else { throw \"SAML endpoint avec le nom '%s' non trouvé\" }",
		name, name,
	)

	var stdout bytes.Buffer
	_, err := client.RunWithContext(ctx, command, &stdout, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la lecture du SAML endpoint : %v", err)
	}

	results := parsePowerShellOutput(stdout.String())

	d.Set("name", results["name"])
	d.Set("relying_party_identifier", results["relying_party_identifier"])
	d.Set("endpoint_url", results["endpoint_url"])
	d.Set("binding", results["binding"])
	d.Set("index", results["index"])

	return nil
}

func resourceAdfsSamlEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	//relyingPartyIdentifier := d.Get("relying_party_identifier").(string)
	endpointUrl := d.Get("endpoint_url").(string)
	binding := d.Get("binding").(string)
	index := d.Get("index").(int)

	command := fmt.Sprintf(
		"$endpoint = Get-AdfsRelyingPartyEndpoint -Name '%s' "+
			"if ($endpoint) { "+
			"$endpoint.Endpoint = '%s'; $endpoint.Binding = '%s'; $endpoint.Index = %d; "+
			"$endpoint | Set-AdfsRelyingPartyEndpoint "+
			"} else { throw \"SAML endpoint avec le nom '%s' non trouvé\" }",
		name, endpointUrl, binding, index, name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la mise à jour du SAML endpoint : %v", err)
	}

	return nil
}

func resourceAdfsSamlEndpointDelete(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"Remove-AdfsRelyingPartyEndpoint -Name '%s'",
		name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la suppression du SAML endpoint : %v", err)
	}

	d.SetId("")

	return nil
}
