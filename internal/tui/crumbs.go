package tui

import "strings"

// crumbText is the breadcrumb bar: the scope, then the views of the stack.
func crumbText(scope string, crumbs []string) string {
	parts := make([]string, 0, len(crumbs)+1)
	parts = append(parts, "[black:aqua] "+plain(orDash(scope))+" [-:-]")
	for i, crumb := range crumbs {
		parts = append(parts, crumbStyle(i == len(crumbs)-1)+" "+plain(crumb)+" [-:-]")
	}
	return strings.Join(parts, " ")
}

// crumbStyle highlights the current view.
func crumbStyle(current bool) string {
	if current {
		return "[black:orange]"
	}
	return "[black:gray]"
}

// scopeLabel names the scope: the public vault, the tenant, or the default tenant.
func (s *shell) scopeLabel() string {
	switch {
	case s.scope.PublicSlug != "":
		return "public:" + s.scope.PublicSlug
	case s.tenantName != "":
		return s.tenantName
	case s.scope.Tenant != "":
		return s.scope.Tenant
	default:
		return "default tenant"
	}
}
