package adfs

import (
	"encoding/json"
	"fmt"
	"github.com/gerardlemetayerc/terraform-provider-adfs/adfs/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os/exec"
	"strings"
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
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"signing_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"endpoints": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"issuance_transform_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_template": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"condition": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type":   {Type: schema.TypeString, Required: true},
									"issuer": {Type: schema.TypeString, Required: true},
								},
							},
						},
						"action": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"store": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"types": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"param": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAdfsRelyingPartyTrustCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	identifier := d.Get("identifier").(string)
	config := m.(*AdfsConfig)
	output, err := helpers.ExecutePowershellCommand(config.PowershellBin, fmt.Sprintf("Add-AdfsRelyingPartyTrust -Name '%s' -Identifier '%s'", name, identifier))
	if err != nil {
		return fmt.Errorf("error creating relying party trust: %s", string(output))
	}

	d.SetId(identifier)

	if err := applyIssuanceTransformRules(d); err != nil {
		return err
	}

	return resourceAdfsRelyingPartyTrustRead(d, m)
}

func resourceAdfsRelyingPartyTrustRead(d *schema.ResourceData, m interface{}) error {
	identifier := d.Id()
	config := m.(*AdfsConfig)

	// Exécuter la commande PowerShell
	output, err := helpers.ExecutePowershellCommand(config.PowershellBin, fmt.Sprintf(
		"Get-AdfsRelyingPartyTrust -Identifier '%s' | Select Name, Identifier, IssuanceTransformRules | ConvertTo-Json -Depth 10", identifier,
	))

	if err != nil {
		if strings.Contains(output, "not found") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading relying party trust: %s", output)
	}

	// Conversion de la sortie en []byte avant l'interprétation JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return fmt.Errorf("error parsing JSON output: %s", err)
	}

	// Remplir les valeurs de la ressource avec les données lues
	if name, ok := result["Name"].(string); ok {
		d.Set("name", name)
	}
	if identifiers, ok := result["Identifier"].([]interface{}); ok && len(identifiers) > 0 {
		d.Set("identifier", identifiers[0])
	}

	// Parser les règles de transformation, si présentes
	if rules, ok := result["IssuanceTransformRules"].(string); ok {
		parsedRules, _ := helpers.ParseIssuanceTransformRules(rules)
		log.Printf("[DEBUG] Parsed issuance transform rules: %+v", parsedRules)
		d.Set("issuance_transform_rules", parsedRules)
	}

	return nil
}

func resourceAdfsRelyingPartyTrustUpdate(d *schema.ResourceData, m interface{}) error {
	if err := executeAdfsCommand("Set-AdfsRelyingPartyTrust", buildAdfsArgs(d)); err != nil {
		return err
	}

	if d.HasChange("issuance_transform_rules") {
		if err := applyIssuanceTransformRules(d); err != nil {
			return err
		}
	}

	return resourceAdfsRelyingPartyTrustRead(d, m)
}

func resourceAdfsRelyingPartyTrustDelete(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)

	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Remove-AdfsRelyingPartyTrust -TargetName '%s'", name))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deleting relying party trust: %s", string(output))
	}
	d.SetId("")
	return nil
}

func buildAdfsArgs(d *schema.ResourceData) string {
	args := fmt.Sprintf("-TargetName '%s' -Identifier '%s'", d.Get("name").(string), d.Get("identifier").(string))

	if signingCertificate, ok := d.GetOk("signing_certificate"); ok {
		args += fmt.Sprintf(" -SigningCertificate '%s'", signingCertificate.(string))
	}

	if endpoints, ok := d.GetOk("endpoints"); ok {
		var endpointStr strings.Builder
		for _, endpoint := range endpoints.([]interface{}) {
			endpointStr.WriteString(fmt.Sprintf("'%s',", endpoint.(string)))
		}
		args += fmt.Sprintf(" -Endpoints @(%s)", strings.TrimSuffix(endpointStr.String(), ","))
	}

	return args
}

func applyIssuanceTransformRules(d *schema.ResourceData) error {
	rules, ok := d.GetOk("issuance_transform_rules")
	if !ok {
		return nil
	}

	ruleStr := buildIssuanceTransformRules(rules.([]interface{}))
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(
		"Set-AdfsRelyingPartyTrust -TargetName '%s' -Identifier '%s' -IssuanceTransformRules '%s'",
		d.Get("name").(string), d.Get("identifier").(string), escapeSingleQuotes(ruleStr),
	))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error applying issuance transform rules: %s", string(output))
	}

	return nil
}

func buildIssuanceTransformRules(rules []interface{}) string {
	var builder strings.Builder

	for _, r := range rules {
		rule := r.(map[string]interface{})
		ruleName := rule["rule_name"].(string)

		builder.WriteString(fmt.Sprintf("@RuleName = \"%s\"\n", ruleName))

		if ruleTemplate, ok := rule["rule_template"].(string); ok && ruleTemplate == "CustomRule" {
			builder.WriteString(rule["rule"].(string))
			builder.WriteString("\n")
			continue
		}

		builder.WriteString(fmt.Sprintf("@RuleTemplate = \"%s\"\n", rule["rule_template"].(string)))

		if condition, ok := rule["condition"].([]interface{}); ok && len(condition) > 0 {
			cond := condition[0].(map[string]interface{})
			builder.WriteString(fmt.Sprintf("c:[Type == \"%s\", Issuer == \"%s\"] ", cond["type"], cond["issuer"]))
		}

		if actions, ok := rule["action"].([]interface{}); ok && len(actions) > 0 {
			builder.WriteString("=> issue(")
			action := actions[0].(map[string]interface{})
			var actionParts []string

			if store, ok := action["store"].(string); ok {
				actionParts = append(actionParts, fmt.Sprintf("store = \"%s\"", store))
			}
			if types, ok := action["types"].([]interface{}); ok {
				var typesStr []string
				for _, t := range types {
					typesStr = append(typesStr, fmt.Sprintf("\"%s\"", t.(string)))
				}
				actionParts = append(actionParts, fmt.Sprintf("types = (%s)", strings.Join(typesStr, ", ")))
			}
			if query, ok := action["query"].(string); ok {
				actionParts = append(actionParts, fmt.Sprintf("query = \"%s\"", query))
			}
			if param, ok := action["param"].(string); ok {
				actionParts = append(actionParts, fmt.Sprintf("param = %s", param))
			}

			builder.WriteString(strings.Join(actionParts, ", "))
			builder.WriteString(");\n")
		}
	}

	return builder.String()
}

func executeAdfsCommand(command string, args string) error {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("%s %s", command, args))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing %s: %s", command, string(output))
	}
	return nil
}

func escapeSingleQuotes(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}
