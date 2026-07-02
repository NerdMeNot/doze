package config

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// ShallowDecl is one engine block header from a shallow load: the type, label,
// and declared engine version — nothing that needs a driver.
type ShallowDecl struct {
	Type    string
	Name    string
	Version string // "" when unset or not statically evaluable (a var reference)
}

// ShallowConfig is the driver-free slice of a project config.
type ShallowConfig struct {
	Modules ModulesConfig
	Decls   []ShallowDecl
	path    string
}

// Path returns the primary config path the shallow load anchored on.
func (c *ShallowConfig) Path() string { return c.path }

// LoadShallow reads just the modules{} block and the engine block headers
// (type, name, declared version) from the config at path — no driver lookup, no
// plugin launch, no body decode, no hooks. It exists for commands that must run
// when the full load can't: `doze modules upgrade` fixes the exact
// protocol/engine-support gates that fail a full load, so it cannot depend on
// one succeeding. Versions referencing variables are returned empty (the module
// requirement is simply not narrowed by them here).
func LoadShallow(path string) (*ShallowConfig, error) {
	files, primary, err := gatherConfigFiles(path)
	if err != nil {
		return nil, err
	}
	parser := hclparse.NewParser()
	hclFiles := make([]*hcl.File, 0, len(files))
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		src, err = rewriteUseBlocks(src, f)
		if err != nil {
			return nil, err
		}
		hf, diags := parser.ParseHCL(src, f)
		if diags.HasErrors() {
			return nil, diagError(parser, diags)
		}
		hclFiles = append(hclFiles, hf)
	}
	merged := hcl.MergeFiles(hclFiles)

	out := &ShallowConfig{path: primary}
	types := engineBlockTypes(hclFiles)
	schema := &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: "modules"}}}
	for _, t := range types {
		schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{Type: t, LabelNames: []string{"name"}})
	}
	content, _, _ := merged.PartialContent(schema)
	for _, block := range content.Blocks {
		if block.Type == "modules" {
			var m hclModules
			// Best-effort: a modules{} block using variables decodes what it can.
			if diags := gohcl.DecodeBody(block.Body, nil, &m); !diags.HasErrors() {
				mc := ModulesConfig{
					Mirror: m.Mirror, Enabled: m.Enabled || m.Mirror != "",
					Sources: map[string]string{}, Versions: map[string]string{},
				}
				for _, md := range m.Modules {
					mc.Sources[md.Name] = md.Source
					mc.Versions[md.Name] = md.Version
				}
				out.Modules = mc
			}
			continue
		}
		decl := ShallowDecl{Type: block.Type}
		if len(block.Labels) == 1 {
			decl.Name = block.Labels[0]
		}
		attrs, _, _ := block.Body.PartialContent(&hcl.BodySchema{Attributes: []hcl.AttributeSchema{{Name: "version"}}})
		if attr, ok := attrs.Attributes["version"]; ok {
			// Static evaluation only — var/local references yield "" rather than
			// dragging the whole evaluator in.
			if v, diags := attr.Expr.Value(nil); !diags.HasErrors() && !v.IsNull() {
				if s, err := convert.Convert(v, cty.String); err == nil && !s.IsNull() {
					decl.Version = s.AsString()
				}
			}
		}
		out.Decls = append(out.Decls, decl)
	}
	return out, nil
}
