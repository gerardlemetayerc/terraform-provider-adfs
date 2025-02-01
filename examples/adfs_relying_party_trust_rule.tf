resource "adfs_relaying_party_trust" "oxy_rp" {
  name       = "MY CUSTOM RP"
  identifier = "MY RP IDENTIFIER"

  issuance_transform_rules {
    rule_template = "LdapClaims"
    rule_name     = "LDAP - Send Attributes"

    condition {
      type   = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname"
      issuer = "AD AUTHORITY"
    }

    action {
      store  = "Active Directory"
      types  = [
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
        "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname",
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
        "http://schemas.xmlsoap.org/claims/Group"
      ]
      query = ";displayName,mail,sAMAccountName,givenName,sn,tokenGroups;{0}"
      param = "c.Value"
    }
  }

  issuance_transform_rules {
    rule_template = "CustomRule"
    rule_name     = "CustomRule - Filter Groups"
    rule          = "c:[Type == \"http://schemas.xmlsoap.org/claims/Group\", Value =~ \"(?i)^GROUP_PATTERN\"] => issue(claim = c);"
  }
}
