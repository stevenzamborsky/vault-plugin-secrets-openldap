package openldap

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"github.com/go-ldap/ldif"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault-plugin-secrets-openldap/client"
)

type ldapClient interface {
	UpdateDNPassword(conf *client.Config, dn string, newPassword string) error
	UpdateUserPassword(conf *client.Config, user, newPassword string) error
	Execute(conf *client.Config, entries []*ldif.Entry, continueOnError bool) error
}

func NewClient(logger hclog.Logger) *Client {
	return &Client{
		ldap: client.New(logger),
	}
}

var _ ldapClient = (*Client)(nil)

type Client struct {
	ldap client.Client
}

// UpdateDNPassword updates the password for the object with the given DN.
func (c *Client) UpdateDNPassword(conf *client.Config, dn string, newPassword string) error {
	filters := map[*client.Field][]string{client.FieldRegistry.ObjectClass: {"*"}}

	newValues, err := client.GetSchemaFieldRegistry(conf.Schema, newPassword)
	if err != nil {
		return fmt.Errorf("error updating password: %s", err)
	}

	return c.ldap.UpdatePassword(conf, dn, ldap.ScopeBaseObject, newValues, filters)
}

// UpdateUserPassword updates the password for the object with the given username.
func (c *Client) UpdateUserPassword(conf *client.Config, username string, newPassword string) error {
	userAttr := conf.UserAttr
	if userAttr == "" {
		userAttr = defaultUserAttr(conf.Schema)
	}

	field := client.FieldRegistry.Parse(userAttr)
	if field == nil {
		return fmt.Errorf("unsupported userattr %q", userAttr)
	}

	filters := map[*client.Field][]string{
		field: {username},
	}

	newValues, err := client.GetSchemaFieldRegistry(conf.Schema, newPassword)
	if err != nil {
		return fmt.Errorf("error updating password: %s", err)
	}

	return c.ldap.UpdatePassword(conf, conf.UserDN, ldap.ScopeWholeSubtree, newValues, filters)
}

func (c *Client) Execute(conf *client.Config, entries []*ldif.Entry, continueOnError bool) (err error) {
	return c.ldap.Execute(conf, entries, continueOnError)
}
