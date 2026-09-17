package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSensitiveDebugLogKeys_CoversEveryTaggedField is the guard that keeps
// sensitiveDebugLogKeys from drifting. The debug logger redacts by JSON key
// name because it works on untyped bodies, so it cannot discover new
// `sensitive:"true"` tags on its own. This test parses the model and config
// sources and fails if any tagged field's JSON name is missing from the set,
// turning a silent leak into a build failure.
func TestSensitiveDebugLogKeys_CoversEveryTaggedField(t *testing.T) {
	tagged := sensitiveJSONNamesInPackages(t, "../models", "../config")
	require.NotEmpty(t, tagged, "found no sensitive:\"true\" tags — the source scan is broken, not the code")

	for name, where := range tagged {
		if _, ok := sensitiveDebugLogKeys[name]; !ok {
			t.Errorf("%s is tagged sensitive:\"true\" but its JSON name %q is missing from sensitiveDebugLogKeys in client.go — --debug would log it in full", where, name)
		}
	}
}

// sensitiveJSONNamesInPackages returns JSON name -> "Type.Field" for every
// struct field tagged `sensitive:"true"` in the given source directories.
func sensitiveJSONNamesInPackages(t *testing.T, dirs ...string) map[string]string {
	t.Helper()
	found := make(map[string]string)

	for _, dir := range dirs {
		packages, err := parser.ParseDir(token.NewFileSet(), dir, func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		require.NoError(t, err, "parsing %s", dir)

		for _, pkg := range packages {
			for _, file := range pkg.Files {
				ast.Inspect(file, func(node ast.Node) bool {
					typeSpec, ok := node.(*ast.TypeSpec)
					if !ok {
						return true
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						return true
					}
					for _, field := range structType.Fields.List {
						if field.Tag == nil {
							continue
						}
						tag, err := strconv.Unquote(field.Tag.Value)
						if err != nil {
							continue
						}
						structTag := reflect.StructTag(tag)
						if structTag.Get("sensitive") != "true" {
							continue
						}
						name := strings.Split(structTag.Get("json"), ",")[0]
						if name == "" || name == "-" {
							continue
						}
						fieldName := name
						if len(field.Names) > 0 {
							fieldName = field.Names[0].Name
						}
						found[name] = typeSpec.Name.Name + "." + fieldName
					}
					return true
				})
			}
		}
	}

	return found
}
