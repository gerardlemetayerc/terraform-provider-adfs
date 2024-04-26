package adfs

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAdfsRelyingPartyTrust() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdfsRelyingPartyTrustCreate,
		Read:   resourceAdfsRelyingPartyTrustRead,
		Update: resourceAdfsRelyingPartyTrustUpdate,
		Delete: resourceAdfsRelyingPartyTrustDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"claims_rules": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAdfsRelyingPartyTrustCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	targetIdentifier := d.Get("target_identifier").(string)
	claimsRules := d.Get("claims_rules").([]string)
	claimsRulesStr := strings.Join(claimsRules, ";")

	command := fmt.Sprintf(
		"New-AdfsRelyingPartyTrust -Name '%s' -Identifier '%s' -ClaimsRules '%s'",
		name, targetIdentifier, claimsRulesStr,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la création du relying party trust : %v", err)
	}

	d.SetId(name)

	return nil
}

func resourceAdfsRelyingPartyTrustRead(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"$trust = Get-AdfsRelyingPartyTrust -Name '%s' "+
			"if ($trust) { "+
			"return @{ 'name' = $trust.Name; 'identifier' = $trust.Identifier; 'claims_rules' = $trust.ClaimsRules -join ';' } "+
			"} else { throw \"Relying party trust avec le nom '%s' non trouvé\" }",
		name, name,
	)

	var stdout bytes.Buffer
	_, err := client.RunWithContext(ctx, command, &stdout, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la lecture du relying party trust : %v", err)
	}

	results := parsePowerShellOutput(stdout.String())

	d.Set("name", results["name"])
	d.Set("target_identifier", results["identifier"])
	d.Set("claims_rules", strings.Split(results["claims_rules"], ";"))

	return nil
}

func resourceAdfsRelyingPartyTrustUpdate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	//targetIdentifier := d.Get("target_identifier").(string)
	claimsRules := d.Get("claims_rules").([]string)
	claimsRulesStr := strings.Join(claimsRules, ";")

	command := fmt.Sprintf(
		"$trust = Get-AdfsRelyingPartyTrust -Name '%s' "+
			"if ($trust) { "+
			"$trust.ClaimsRules = '%s'.split(';'); "+
			"$trust | Set-AdfsRelyingPartyTrust "+
			"} else { throw \"Relying party trust avec le nom '%s' non trouvé\" }",
		name, claimsRulesStr, name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la mise à jour du relying party trust : %v", err)
	}

	return nil
}

func resourceAdfsRelyingPartyTrustDelete(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"Remove-AdfsRelyingPartyTrust -Name '%s'",
		name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("Erreur lors de la suppression du relying party trust : %v", err)
	}

	d.SetId("")

	return nil
}
