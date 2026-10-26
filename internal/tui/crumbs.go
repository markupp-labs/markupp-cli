package tui

// crumbText is the breadcrumb bar: the scope, then the views of the stack.
func crumbText(scope string, crumbs []string) string { return *new(string) }

// crumbStyle highlights the current view.
func crumbStyle(current bool) string { return *new(string) }

// scopeLabel names the scope: the public vault, the tenant, or the default tenant.
func (s *shell) scopeLabel() string { return *new(string) }
