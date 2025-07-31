//coverage:ignore file

// This generator is mostly written by AI, due to lack of better help templating solution
// Not great, not terrible :harold:

// TODO: replace with better flag-parsing library having support for flag grouping
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"sort"
	"strings"
)

const (
	srcFlagFile = "exporter_cmd.go"
)

type Flag struct {
	Name        string
	Description string
	Default     string
	Type        string
}

type Command struct {
	Name        string
	Description string
	Flags       []Flag
}

// extractCommandInfo parses an ast.Expr to find kingpin.Command calls.
// It returns the command name and description if found.
func extractCommandInfo(expr ast.Expr) (name, description string) {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return
	}

	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok || selectorExpr.Sel.Name != "Command" {
		return
	}

	// Ensure it's kingpin.Command
	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok || ident.Name != "kingpin" {
		return
	}

	// Extract arguments: name and description
	if len(callExpr.Args) >= 2 {
		if nameLit, ok := callExpr.Args[0].(*ast.BasicLit); ok && nameLit.Kind == token.STRING {
			name = strings.Trim(nameLit.Value, `"`)
		}
		if descLit, ok := callExpr.Args[1].(*ast.BasicLit); ok && descLit.Kind == token.STRING {
			description = strings.Trim(descLit.Value, `"`)
		}
	}
	return
}

// extractFlagInfo parses an ast.Expr to find kingpin.Flag or command.Flag calls.
// It traverses the chained calls (.Default(), .String(), etc.) to extract all relevant info.
func extractFlagInfo(expr ast.Expr) (name, description, defaultValue, flagType string, isCommandFlag bool, cmdVarName string) {
	// Start from the outermost call (e.g., .String(), .Bool(), .ExistingFile())
	currentExpr := expr
	extraDescription := ""

	// Traverse backwards to find the type method and default value
	for {
		callExpr, ok := currentExpr.(*ast.CallExpr)
		if !ok {
			break // Not a call expression, stop traversing
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			break // Not a selector expression (e.g., kingpin.Flag), stop traversing
		}

		methodName := selectorExpr.Sel.Name

		switch methodName {
		case "String", "Bool", "Duration", "Regexp", "ExistingFile", "ExistingDir":
			flagType = methodName
		case "Enum":
			enumValues := []string{}
			for _, arg := range callExpr.Args {
				if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					enumValues = append(enumValues, strings.Trim(lit.Value, `"`))
				}
			}
			extraDescription = fmt.Sprintf(". Possible values are: %s", strings.Join(enumValues, ", "))
		case "Default":
			if len(callExpr.Args) > 0 {
				if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					defaultValue = strings.Trim(lit.Value, `"`)
				}
			}
		case "Flag":
			// This is the kingpin.Flag or cmd.Flag call itself
			// Check if it's a command flag (e.g., `cmdVar.Flag`) or global (`kingpin.Flag`)
			if ident, ok := selectorExpr.X.(*ast.Ident); ok {
				if ident.Name != "kingpin" {
					isCommandFlag = true
					cmdVarName = ident.Name
				}
			}

			if len(callExpr.Args) >= 2 {
				if nameLit, ok := callExpr.Args[0].(*ast.BasicLit); ok && nameLit.Kind == token.STRING {
					name = strings.Trim(nameLit.Value, `"`)
				}
				if descLit, ok := callExpr.Args[1].(*ast.BasicLit); ok && descLit.Kind == token.STRING {
					description = strings.Trim(descLit.Value, `"`)
				}
			}
			description += extraDescription
			return // Found the Flag call, we have all info
		default:
			slog.Error("Skipping unknown kingpin method", "methodName", methodName)
		}
		currentExpr = selectorExpr.X // Move to the receiver of the current method call
	}
	description += extraDescription
	return
}

func formatFlag(f Flag, indent int) string {
	indentString := strings.Repeat(" ", indent)
	desc := ""
	if f.Description != "" {
		doubleIndentString := strings.Repeat(" ", indent*2)
		desc = fmt.Sprintf("\n%s%s", doubleIndentString, f.Description)
	}
	if f.Type == "Bool" {
		if f.Default == "true" {
			return fmt.Sprintf("%s--no-%s%s", indentString, f.Name, desc)
		}
		return fmt.Sprintf("%s--%s%s", indentString, f.Name, desc)
	}
	return fmt.Sprintf("%s--%s=%s%s", indentString, f.Name, f.Default, desc)
}

func main() {
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, srcFlagFile, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Grouped global flags
	groups := map[string][]string{}
	groupOrder := []string{}
	seenGroups := map[string]bool{}
	currentGroup := "Ungrouped"

	// Commands and their flags
	commands := map[string]*Command{}
	cmdVars := map[string]string{} // map var name to command name

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Doc comments for flag grouping (this part remains the same)
			if valueSpec.Doc != nil {
				for _, c := range valueSpec.Doc.List {
					txt := strings.TrimSpace(c.Text)
					if after, ok := strings.CutPrefix(txt, "// FLAG GROUP START:"); ok {
						currentGroup = strings.TrimSpace(after)
						if !seenGroups[currentGroup] {
							groupOrder = append(groupOrder, currentGroup)
							seenGroups[currentGroup] = true
						}
					}
					if strings.HasPrefix(txt, "// FLAG GROUP END") {
						currentGroup = "Ungrouped"
					}
				}
			}

			if len(valueSpec.Values) == 0 {
				continue
			}

			// Directly inspect the AST expression instead of converting to string and using regex
			expr := valueSpec.Values[0]

			// Try to extract command info
			cmdName, cmdDesc := extractCommandInfo(expr)
			if cmdName != "" {
				varName := valueSpec.Names[0].Name
				commands[cmdName] = &Command{Name: cmdName, Description: cmdDesc}
				cmdVars[varName] = cmdName
				continue
			}

			// Try to extract flag info
			flagName, flagDesc, flagDefault, flagType, isCmdFlag, cmdVar := extractFlagInfo(expr)
			if flagName != "" {
				if isCmdFlag {
					if cmd, ok := commands[cmdVars[cmdVar]]; ok {
						cmd.Flags = append(cmd.Flags, Flag{flagName, flagDesc, flagDefault, flagType})
					}
				} else {
					// Global flag
					if !seenGroups[currentGroup] {
						groupOrder = append(groupOrder, currentGroup)
						seenGroups[currentGroup] = true
					}
					groups[currentGroup] = append(groups[currentGroup], formatFlag(Flag{flagName, flagDesc, flagDefault, flagType}, 2))
				}
				continue
			}
		}
	}

	// Final help text build
	var helpBuffer bytes.Buffer
	helpBuffer.WriteString("`")
	helpBuffer.WriteString("USAGE:\n")

	// Commands
	if len(commands) > 0 {
		helpBuffer.WriteString("Commands:\n\n")
		var cmdNames []string
		for name := range commands {
			cmdNames = append(cmdNames, name)
		}
		sort.Strings(cmdNames)

		for _, name := range cmdNames {
			cmd := commands[name]
			helpBuffer.WriteString(name + ":\n")
			if cmd.Description != "" {
				helpBuffer.WriteString("  " + cmd.Description + "\n")
			}
			for _, f := range cmd.Flags {
				helpBuffer.WriteString(formatFlag(f, 4) + "\n\n")
			}
		}

	}

	helpBuffer.WriteString("\n\nFlags:\n\n")
	// Grouped global flags
	for _, group := range groupOrder {
		helpBuffer.WriteString(group + ":\n")
		for _, line := range groups[group] {
			helpBuffer.WriteString(line + "\n")
		}
		helpBuffer.WriteString("\n")
	}
	helpBuffer.WriteString("`")

	out := fmt.Sprintf(`// THIS FILE IS AUTO-GENERATED BY utils/generate_exporter_help.go. DO NOT EDIT.

package main

const helpText = %s
`, helpBuffer.String())

	os.Stdout.WriteString(out)
}
