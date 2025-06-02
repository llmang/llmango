package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseGoFiles scans Go files in the given directory and extracts goal and prompt definitions
func ParseGoFiles(dir string) (*ParseResult, error) {
	result := &ParseResult{
		Goals:   []DiscoveredGoal{},
		Prompts: []DiscoveredPrompt{},
		Errors:  []ParseError{},
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go files: %w", err)
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			relPath, _ := filepath.Rel(dir, filename)
			parseFile(fset, file, relPath, result)
		}
	}

	return result, nil
}

// parseFile extracts goals and prompts from a single Go file
func parseFile(fset *token.FileSet, file *ast.File, filename string, result *ParseResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						parseValueSpec(fset, valueSpec, filename, result)
					}
				}
			}
		}
		return true
	})
}

// parseValueSpec extracts goal or prompt from a variable declaration
func parseValueSpec(fset *token.FileSet, spec *ast.ValueSpec, filename string, result *ParseResult) {
	for i, name := range spec.Names {
		if i < len(spec.Values) {
			value := spec.Values[i]
			
			// Check if this is a goal or prompt definition
			if compositeLit, ok := value.(*ast.CompositeLit); ok {
				if selectorExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok {
					if ident, ok := selectorExpr.X.(*ast.Ident); ok && ident.Name == "llmango" {
						switch selectorExpr.Sel.Name {
						case "Goal":
							goal := parseGoalLiteral(fset, compositeLit, name.Name, filename)
							if goal != nil {
								result.Goals = append(result.Goals, *goal)
							}
						case "Prompt":
							prompt := parsePromptLiteral(fset, compositeLit, name.Name, filename)
							if prompt != nil {
								result.Prompts = append(result.Prompts, *prompt)
							}
						}
					}
				}
			}
		}
	}
}

// parseGoalLiteral extracts goal information from a composite literal
func parseGoalLiteral(fset *token.FileSet, lit *ast.CompositeLit, varName, filename string) *DiscoveredGoal {
	goal := &DiscoveredGoal{
		VarName:    varName,
		SourceFile: filename,
		SourceType: "go",
	}

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "UID":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						goal.UID = strings.Trim(basicLit.Value, `"`)
					}
				case "Title":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						goal.Title = strings.Trim(basicLit.Value, `"`)
					}
				case "Description":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						goal.Description = strings.Trim(basicLit.Value, `"`)
					}
				case "InputOutput":
					// Extract input/output types from InputOutput[I, R] generic
					if compositeLit, ok := kv.Value.(*ast.CompositeLit); ok {
						// Handle both IndexExpr (Go 1.18+) and IndexListExpr
						switch typeExpr := compositeLit.Type.(type) {
						case *ast.IndexExpr:
							if selectorExpr, ok := typeExpr.X.(*ast.SelectorExpr); ok {
								if ident, ok := selectorExpr.X.(*ast.Ident); ok && ident.Name == "llmango" && selectorExpr.Sel.Name == "InputOutput" {
									// For single type parameter (shouldn't happen with InputOutput but handle it)
									if inputIdent, ok := typeExpr.Index.(*ast.Ident); ok {
										goal.InputType = inputIdent.Name
									}
								}
							}
						case *ast.IndexListExpr:
							if selectorExpr, ok := typeExpr.X.(*ast.SelectorExpr); ok {
								if ident, ok := selectorExpr.X.(*ast.Ident); ok && ident.Name == "llmango" && selectorExpr.Sel.Name == "InputOutput" {
									// Extract generic type parameters
									if len(typeExpr.Indices) >= 2 {
										if inputIdent, ok := typeExpr.Indices[0].(*ast.Ident); ok {
											goal.InputType = inputIdent.Name
										}
										if outputIdent, ok := typeExpr.Indices[1].(*ast.Ident); ok {
											goal.OutputType = outputIdent.Name
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Validate required fields
	if goal.UID == "" || goal.InputType == "" || goal.OutputType == "" {
		return nil
	}

	return goal
}

// parsePromptLiteral extracts prompt information from a composite literal
func parsePromptLiteral(fset *token.FileSet, lit *ast.CompositeLit, varName, filename string) *DiscoveredPrompt {
	prompt := &DiscoveredPrompt{
		VarName:    varName,
		SourceFile: filename,
		SourceType: "go",
		Weight:     100, // Default weight
	}

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "UID":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						prompt.UID = strings.Trim(basicLit.Value, `"`)
					}
				case "GoalUID":
					// Handle both string literals and variable references
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						prompt.GoalUID = strings.Trim(basicLit.Value, `"`)
					} else if selectorExpr, ok := kv.Value.(*ast.SelectorExpr); ok {
						// Handle cases like testGoal.UID
						if ident, ok := selectorExpr.X.(*ast.Ident); ok && selectorExpr.Sel.Name == "UID" {
							// For now, we'll extract the goal UID from the variable name
							// This is a simplified approach - in a real parser we'd need to resolve the variable
							goalVarName := ident.Name
							// Convert variable name to likely UID (remove "Goal" suffix if present)
							if strings.HasSuffix(goalVarName, "Goal") {
								goalVarName = goalVarName[:len(goalVarName)-4]
							}
							// Convert camelCase to kebab-case
							prompt.GoalUID = convertCamelToKebab(goalVarName)
						}
					}
				case "Model":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
						prompt.Model = strings.Trim(basicLit.Value, `"`)
					}
				case "Weight":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
						// Parse weight (simplified - would need proper int parsing)
						prompt.Weight = 100 // Default for now
					}
				case "IsCanary":
					if ident, ok := kv.Value.(*ast.Ident); ok {
						prompt.IsCanary = ident.Name == "true"
					}
				case "MaxRuns":
					if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
						// Parse max runs (simplified)
						prompt.MaxRuns = 0 // Default for now
					}
				}
			}
		}
	}

	// Validate required fields
	if prompt.UID == "" || prompt.GoalUID == "" || prompt.Model == "" {
		return nil
	}

	return prompt
}

// convertCamelToKebab converts camelCase to kebab-case
func convertCamelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GenerateMethodName converts a goal UID to a valid Go method name
func GenerateMethodName(goalUID string) string {
	// Remove non-alphanumeric characters and convert to PascalCase
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	parts := re.Split(goalUID, -1)
	
	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			result.WriteString(strings.Title(strings.ToLower(part)))
		}
	}
	
	methodName := result.String()
	if methodName == "" {
		methodName = "UnnamedGoal"
	}
	
	return methodName
}