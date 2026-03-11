package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type UniqueRule struct {
	Checker     UniquenessChecker
	CtxProvider func() context.Context
}

func (u UniqueRule) Register(v *validator.Validate) error {
	return v.RegisterValidation("unique", u.unique)
}

type uniqueParam struct {
	Table        string
	Column       string
	ScopeFields  []string
	IgnoreFields []string
}

func (u UniqueRule) unique(fl validator.FieldLevel) bool {
	param := fl.Param()
	cfg, err := parseUniqueParam(param)
	if err != nil {
		return false
	}

	val := fl.Field().Interface()
	parent := fl.Parent()

	scope := map[string]any{}
	for _, name := range cfg.ScopeFields {
		f := parent.FieldByName(name)
		if !f.IsValid() {
			return false
		}
		scope[snake(name)] = f.Interface()
	}

	ignore := map[string]any{}
	for _, name := range cfg.IgnoreFields {
		f := parent.FieldByName(name)
		if !f.IsValid() {
			return false
		}
		ignore[snake(name)] = f.Interface()
	}

	ctx := context.Background()
	if u.CtxProvider != nil {
		if c := u.CtxProvider(); c != nil {
			ctx = c
		}
	}

	ok, err := u.Checker.IsUnique(ctx, cfg.Table, cfg.Column, val, scope, ignore)
	return err == nil && ok
}

func parseUniqueParam(p string) (uniqueParam, error) {
	var res uniqueParam
	parts := strings.Split(p, ",")
	if len(parts) == 0 || !strings.Contains(parts[0], ".") {
		return res, fmt.Errorf("unique: expected table.column")
	}
	tc := strings.SplitN(parts[0], ".", 2)
	res.Table, res.Column = tc[0], tc[1]

	for _, s := range parts[1:] {
		if strings.HasPrefix(s, "scope=") {
			res.ScopeFields = splitCSV(strings.TrimPrefix(s, "scope="))
		} else if strings.HasPrefix(s, "ignore=") {
			res.IgnoreFields = splitCSV(strings.TrimPrefix(s, "ignore="))
		}
	}
	return res, nil
}
func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, "|") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func snake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
