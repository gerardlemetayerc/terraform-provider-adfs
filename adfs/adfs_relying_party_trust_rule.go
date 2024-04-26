package adfs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAdfsRelyingPartyTrustRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdfsRelyingPartyTrustRuleCreate,
		Read:   resourceAdfsRelyingPartyTrustRuleRead,
		Update: resourceAdfsRelyingPartyTrustRuleUpdate,
		Delete: resourceAdfsRelyingPartyTrustRuleDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_language": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAdfsRelyingPartyTrustRuleCreate(d *schema.ResourceData, m interface{}) error {
	var diags diag.Diagnostics
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	targetIdentifier := d.Get("target_identifier").(string)
	ruleLanguage := d.Get("rule_language").(string)
	ruleType := d.Get("rule_type").(string)
	rule := d.Get("rule").(string)

	command := fmt.Sprintf(
		"New-AdfsRelyingPartyTrustRule -Name '%s' -TargetIdentifier '%s' -RuleLanguage '%s' -RuleType '%s' -Rule '%s'",
		name, targetIdentifier, ruleLanguage, ruleType, rule,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la création de la règle d'émission",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	if len(diags) == 0 {
		d.SetId(name)
	}

	return nil
}

func resourceAdfsRelyingPartyTrustRuleRead(d *schema.ResourceData, m interface{}) error {
	var diags diag.Diagnostics
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"$rule = Get-AdfsRelyingPartyTrustRule -Name '%s' "+
			"if ($rule) { "+
			"return @{ 'name' = $rule.Name; 'target_identifier' = $rule.TargetIdentifier; "+
			"'rule_language' = $rule.RuleLanguage; 'rule_type' = $rule.RuleType; 'rule' = $rule.Rule } "+
			"} else { throw \"Règle avec le nom '%s' non trouvée\" }",
		name, name,
	)

	var stdout bytes.Buffer
	_, err := client.RunWithContext(ctx, command, &stdout, &bytes.Buffer{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la lecture de la règle d'émission",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	results := parsePowerShellOutput(stdout.String())

	d.Set("name", results["name"])
	d.Set("target_identifier", results["target_identifier"])
	d.Set("rule_language", results["rule_language"])
	d.Set("rule_type", results["rule_type"])
	d.Set("rule", results["rule"])

	return nil
}

func resourceAdfsRelyingPartyTrustRuleUpdate(d *schema.ResourceData, m interface{}) error {
	var diags diag.Diagnostics
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)
	//targetIdentifier := d.Get("target_identifier").(string)
	ruleLanguage := d.Get("rule_language").(string)
	ruleType := d.Get("rule_type").(string)
	rule := d.Get("rule").(string)

	command := fmt.Sprintf(
		"$rule = Get-AdfsRelyingPartyTrustRule -Name '%s' "+
			"if ($rule) { "+
			"$rule.RuleLanguage = '%s'; $rule.RuleType = '%s'; $rule.Rule = '%s'; "+
			"$rule | Set-AdfsRelyingPartyTrustRule "+
			"} else { throw \"Règle avec le nom '%s' non trouvée\" }",
		name, ruleLanguage, ruleType, rule, name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la mise à jour de la règle d'émission",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	return nil
}

func resourceAdfsRelyingPartyTrustRuleDelete(d *schema.ResourceData, m interface{}) error {
	var diags diag.Diagnostics
	ctx := context.Background()

	config := m.(*AdfsConfig)
	client := config.Client

	name := d.Get("name").(string)

	command := fmt.Sprintf(
		"Remove-AdfsRelyingPartyTrustRule -Name '%s'",
		name,
	)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la suppression de la règle d'émission",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	d.SetId("")
	return nil
}
