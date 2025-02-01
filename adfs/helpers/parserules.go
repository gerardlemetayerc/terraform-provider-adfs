package helpers

import (
	"log"
	"strings"
)

func ParseIssuanceTransformRules(rules string) ([]map[string]interface{}, error) {
	var parsedRules []map[string]interface{}
	ruleBlocks := strings.Split(rules, "\r\n\r\n")

	for _, block := range ruleBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			log.Printf("[DEBUG] Skipping empty rule block.")
			continue
		}

		log.Printf("[DEBUG] Parsing rule block: %s", block)
		rule := make(map[string]interface{})
		lines := strings.Split(block, "\r\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			log.Printf("[DEBUG] Parsing line: %s", line)

			// @RuleName
			if strings.HasPrefix(line, "@RuleName =") {
				ruleName := strings.Trim(strings.SplitN(line, "=", 2)[1], "\" ")
				rule["rule_name"] = ruleName
				log.Printf("[DEBUG] Found rule_name: %s", ruleName)

				// @RuleTemplate
			} else if strings.HasPrefix(line, "@RuleTemplate =") {
				ruleTemplate := strings.Trim(strings.SplitN(line, "=", 2)[1], "\" ")
				rule["rule_template"] = ruleTemplate
				log.Printf("[DEBUG] Found rule_template: %s", ruleTemplate)

				// Condition (starts with c:)
			} else if strings.HasPrefix(line, "c:[") {
				conditionContent := strings.TrimSuffix(strings.TrimPrefix(line, "c:["), "]")
				log.Printf("[DEBUG] Found condition block: %s", conditionContent)

				conditionParts := strings.Split(conditionContent, ", ")
				conditions := make(map[string]string)
				for _, part := range conditionParts {
					kv := strings.SplitN(part, "==", 2)
					if len(kv) == 2 {
						key := strings.TrimSpace(kv[0])
						value := strings.Trim(strings.TrimSpace(kv[1]), "\"")
						conditions[key] = value
						log.Printf("[DEBUG] Parsed condition: %s == %s", key, value)
					}
				}
				rule["condition"] = conditions

				// Action (starts with => issue)
			} else if strings.HasPrefix(line, "=> issue(") {
				actionContent := strings.TrimSuffix(strings.TrimPrefix(line, "=> issue("), ");")
				log.Printf("[DEBUG] Found action block: %s", actionContent)

				actionParts := strings.Split(actionContent, ", ")
				actions := make(map[string]interface{})
				for _, part := range actionParts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						key := strings.TrimSpace(kv[0])
						value := strings.Trim(strings.TrimSpace(kv[1]), "\"")

						// Handle 'types' as a list
						if key == "types" {
							value = strings.Trim(value, "()")
							actions[key] = strings.Split(value, "\", \"")
							log.Printf("[DEBUG] Parsed action types: %v", actions[key])
						} else {
							actions[key] = value
							log.Printf("[DEBUG] Parsed action %s: %s", key, value)
						}
					}
				}
				rule["action"] = actions
			}
		}

		// Handle default case when no rule_template is present (assume CustomRule)
		if _, ok := rule["rule_template"]; !ok {
			rule["rule_template"] = "CustomRule"
			rule["rule"] = block
			log.Printf("[DEBUG] No rule_template found, assuming CustomRule.")
		}

		log.Printf("[DEBUG] Successfully parsed rule: %+v", rule)
		parsedRules = append(parsedRules, rule)
	}

	log.Printf("[DEBUG] Completed parseIssuanceTransformRules. Parsed rules: %+v", parsedRules)
	return parsedRules, nil
}
