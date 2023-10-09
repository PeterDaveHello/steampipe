package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type ParsedPropertyPath struct {
	Mod          string
	ItemType     string
	Name         string
	PropertyPath []string
	// optional scope of this property path ("self")
	Scope    string
	Original string
}

func (p *ParsedPropertyPath) PropertyPathString() string {
	return strings.Join(p.PropertyPath, ".")
}

func (p *ParsedPropertyPath) ToParsedResourceName() *ParsedResourceName {
	return &ParsedResourceName{
		Mod:      p.Mod,
		ItemType: p.ItemType,
		Name:     p.Name,
	}
}

func (p *ParsedPropertyPath) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedPropertyPath) String() string {
	return p.Original
}

func ParseResourcePropertyPath(propertyPath string, hclRange hcl.Range) (*ParsedPropertyPath, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := &ParsedPropertyPath{Original: propertyPath}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type

	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("invalid property path '%s' passed to ParseResourcePropertyPath", propertyPath),
			Subject: &hclRange,
		})
		return nil, diags
	}

	// special case handling for runtime dependencies which may have use the "self" qualifier
	if parts[0] == RuntimeDependencyDashboardScope {
		res.Scope = parts[0]
		parts = parts[1:]
	}

	if IsValidResourceItemType(parts[0]) {
		// put empty mod as first part - so we can assume always that the first part is mod
		parts = append([]string{""}, parts...)
	}
	// at this point the length of property path must be at least 3 (i.e.<mod>.<resource>.<name>)
	if len(parts) < 3 {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("invalid property path '%s' passed to ParseResourcePropertyPath", propertyPath),
			Subject: &hclRange,
		})
		return nil, diags
	}
	switch len(parts) {
	case 3:
		// no property path specified
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
	default:
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
		res.PropertyPath = parts[3:]
	}

	if !IsValidResourceItemType(res.ItemType) {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("invalid property path '%s' passed to ParseResourcePropertyPath", propertyPath),
			Subject: &hclRange,
		})
		return nil, diags
	}

	return res, nil
}
