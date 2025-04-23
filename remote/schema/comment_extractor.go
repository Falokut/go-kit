// nolint
package schema

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// ExtractGoComments reads all Go files in a given path and its subdirectories,
// extracting doc comments for exported types and their fields.
// `base` is the import path prefix to match the final FQNs.
func ExtractGoComments(base string, rootPath string, commentMap map[string]string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, filepath.Join(rootPath, "..."))
	if err != nil {
		return errors.WithMessagef(err, "failed to load packages from directory %s", rootPath)
	}

	for _, pkg := range pkgs {
		if pkg.PkgPath == "" {
			continue
		}
		importPath := filepath.ToSlash(filepath.Join(base, strings.TrimPrefix(pkg.PkgPath, rootPath)))

		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Doc != nil {
						typeKey := fmt.Sprintf("%s.%s", importPath, typeSpec.Name.Name)
						commentMap[typeKey] = strings.TrimSpace(typeSpec.Doc.Text())
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					for _, field := range structType.Fields.List {
						txt := field.Doc.Text()
						if txt == "" {
							txt = field.Comment.Text()
						}
						if txt == "" {
							continue
						}
						for _, name := range field.Names {
							if ast.IsExported(name.Name) {
								fieldKey := fmt.Sprintf("%s.%s.%s", importPath, typeSpec.Name.Name, name.Name)
								commentMap[fieldKey] = strings.TrimSpace(txt)
							}
						}
					}
				}
			}
		}
	}

	return nil
}
