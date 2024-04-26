package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Ressource Terraform pour AdfsClient
func resourceAdfsClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdfsClientCreate,
		Read:   resourceAdfsClientRead,
		Update: resourceAdfsClientUpdate,
		Delete: resourceAdfsClientDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nom du client",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID unique pour le client",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Secret pour le client",
			},
			"redirect_uris": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "Liste des 'redirect URIs'",
			},
			"scopes": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "Liste des 'scopes' pour le client",
			},
		},
	}
}

// Fonction pour créer un AdfsClient
func resourceAdfsClientCreate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	var diags diag.Diagnostics

	config := m.(*AdfsConfig) // Récupère la configuration du provider
	client := config.Client   // Utilise le client WinRM global

	name := d.Get("name").(string)
	clientId := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	redirectUris := d.Get("redirect_uris").([]string)
	redirectUrisStr := strings.Join(redirectUris, ";") // Regroupe les "redirect URIs"

	command := fmt.Sprintf(`
        $redirectUris = '%s'.split(';')
        New-AdfsClient -Name '%s' -ClientId '%s' -RedirectUri $redirectUris -ClientSecret '%s'
    `, redirectUrisStr, name, clientId, clientSecret)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{}) // Exécution avec RunWithContext
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la création du AdfsClient",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
	}

	if len(diags) == 0 {
		d.SetId(clientId) // Définit l'identifiant si l'opération réussit
	}

	return nil
}

// Fonction pour lire les détails d'un AdfsClient
func resourceAdfsClientRead(d *schema.ResourceData, m interface{}) error {
	var diags diag.Diagnostics
	ctx := context.Background()

	config := m.(*AdfsConfig) // Récupère la configuration
	client := config.Client   // Utilise le client WinRM global

	clientId := d.Id() // Identifiant du client à lire

	command := fmt.Sprintf(`
        $client = Get-AdfsClient -ClientId '%s'
        if ($client) {
            $redirectUris = $client.RedirectUris -join ';'
            return @{
                'name' = $client.Name
                'client_id' = $client.ClientId
                'redirect_uris' = $redirectUris
            }
        } else {
            throw "Client with ID '%s' not found"
        }
    `, clientId, clientId)

	var stdout bytes.Buffer
	_, err := client.RunWithContext(ctx, command, &stdout, &bytes.Buffer{}) // Exécution avec RunWithContext
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la lecture du AdfsClient",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	// Parsing des résultats de la commande PowerShell
	results := parsePowerShellOutput(stdout.String())

	// Mise à jour de l'état Terraform avec les détails obtenus
	d.Set("name", results["name"])
	d.Set("client_id", results["client_id"])
	d.Set("redirect_uris", strings.Split(results["redirect_uris"], ";")) // Retourne les "redirect URIs" comme tableau

	return nil
}

// Fonction pour mettre à jour un AdfsClient
func resourceAdfsClientUpdate(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	var diags diag.Diagnostics

	config := m.(*AdfsConfig) // Récupère la configuration du provider
	client := config.Client   // Utilise le client WinRM global

	clientId := d.Get("client_id").(string)           // Identifiant du client
	redirectUris := d.Get("redirect_uris").([]string) // Mise à jour des "redirect URIs"
	redirectUrisStr := strings.Join(redirectUris, ";")

	command := fmt.Sprintf(`
        $client = Get-AdfsClient -ClientId '%s'
        if ($client) {
            $client.RedirectUris = '%s'.split(';')
            $client | Set-AdfsClient
        } else {
            throw "Client with ID '%s' not found"
        }
    `, clientId, redirectUrisStr, clientId)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{}) // Utilisation de RunWithContext
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la mise à jour du AdfsClient",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	return nil // Retourne les diagnostics pour indiquer le succès ou les erreurs
}

// Fonction pour supprimer un AdfsClient
func resourceAdfsClientDelete(d *schema.ResourceData, m interface{}) error {
	ctx := context.Background()
	var diags diag.Diagnostics

	config := m.(*AdfsConfig) // Récupère la configuration
	client := config.Client   // Utilise le client WinRM global

	clientId := d.Id() // Identifiant du client à supprimer

	command := fmt.Sprintf(`
        Remove-AdfsClient -ClientId '%s'
    `, clientId)

	_, err := client.RunWithContext(ctx, command, &bytes.Buffer{}, &bytes.Buffer{}) // Exécution avec RunWithContext
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Erreur lors de la suppression du AdfsClient",
			Detail:   fmt.Sprintf("Erreur lors de l'exécution de la commande PowerShell: %s", err.Error()),
		})
		return err
	}

	d.SetId("") // Efface l'ID si l'opération réussit

	return nil // Retourne les diagnostics pour indiquer le succès ou les erreurs
}
