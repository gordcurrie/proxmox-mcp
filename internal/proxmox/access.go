package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListUsers returns all users configured in the Proxmox access control
// system, across all realms (pve, pam, ldap, etc). Useful for auditing for
// unrecognized or orphaned accounts.
func (c *Client) ListUsers(ctx context.Context) ([]map[string]any, error) {
	var users []map[string]any
	if err := c.get(ctx, "/access/users", &users); err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return users, nil
}

// ListUserTokens returns the API tokens issued to a single user. Token
// secrets are never included by the Proxmox API — only metadata (token ID,
// comment, expiry, privilege separation).
func (c *Client) ListUserTokens(ctx context.Context, userid string) ([]map[string]any, error) {
	var tokens []map[string]any
	if err := c.get(ctx, "/access/users/"+url.PathEscape(userid)+"/token", &tokens); err != nil {
		return nil, fmt.Errorf("listing tokens for user %s: %w", userid, err)
	}
	return tokens, nil
}
