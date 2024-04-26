package main

import (
	"regexp"
)

// Fonction pour analyser la sortie de PowerShell et retourner une map clé-valeur
func parsePowerShellOutput(output string) map[string]string {
	results := make(map[string]string)

	// Exemple d'utilisation d'une expression régulière pour extraire des paires clé-valeur
	// Suppose que la sortie est sous la forme "clé = valeur"
	regex := regexp.MustCompile(`'(?P<key>\w+)' = '(?P<value>.+?)'`)
	matches := regex.FindAllStringSubmatch(output, -1)

	if matches != nil {
		for _, match := range matches {
			key := match[1]
			value := match[2]
			results[key] = value
		}
	}

	return results // Retourne la map des paires clé-valeur
}
