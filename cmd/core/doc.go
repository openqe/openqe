package core

import (
	"encoding/json"
	"fmt"
	"log"

	godoc "go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewDocCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "doc",
		Short:        "Documentation related commands",
		SilenceUsage: true,
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	cmd.AddCommand(NewCobraDocGenCmd(rootCmd))
	cmd.AddCommand(NewDocDumpCmd(rootCmd))
	return cmd
}

type DocGenOptions struct {
	Output string
}

var opts = &DocGenOptions{}

func NewCobraDocGenCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cobra-doc-gen",
		Short:        "Generate the markdown documentation for the CLI",
		SilenceUsage: true,
	}
	cmd.Flags().StringVar(&opts.Output, "output", opts.Output, "The CA certificate subject used to generate the TLS CA.")
	cmd.MarkFlagRequired("output")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(rootCmd, opts.Output)
		if err != nil {
			return fmt.Errorf("Failed to generate the documentation: %v\n", err)
		}
		cmd.Printf("The documentation is generated to %s\n", opts.Output)
		return nil
	}
	return cmd
}

// API docs structure for JSON
type APIDocs struct {
	Package   string     `json:"package,omitempty"`
	Functions []FuncDoc  `json:"functions,omitempty"`
	Types     []TypeDoc  `json:"types,omitempty"`
	Consts    []ValueDoc `json:"consts,omitempty"`
	Vars      []ValueDoc `json:"vars,omitempty"`
}

type FuncDoc struct {
	Name      string `json:"name,omitempty"`
	Signature string `json:"signature,omitempty"`
	Doc       string `json:"doc,omitempty"`
}

type MethodDoc struct {
	Name      string `json:"name,omitempty"`
	Signature string `json:"signature,omitempty"`
	Doc       string `json:"doc,omitempty"`
}

type TypeDoc struct {
	Name    string      `json:"name,omitempty"`
	Doc     string      `json:"doc,omitempty"`
	Methods []MethodDoc `json:"methods,omitempty"`
}

type ValueDoc struct {
	Names []string `json:"names,omitempty"`
	Doc   string   `json:"doc,omitempty"`
}

type DocDumpOptions struct {
	ProjectBaseDir string // the base directory of the Go project
	OutDir         string // the output directory for the JSON files
}

func NewDocDumpCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "doc-dump",
		Short:        "Dump public API docs of a Go project to JSON format",
		SilenceUsage: true,
	}
	docDumpOpts := &DocDumpOptions{}
	cmd.Flags().StringVar(&docDumpOpts.ProjectBaseDir, "project-base-dir", docDumpOpts.ProjectBaseDir, "The base directory of the Go project.")
	cmd.MarkFlagRequired("project-base-dir")
	cmd.Flags().StringVar(&docDumpOpts.OutDir, "out-dir", docDumpOpts.OutDir, "The output directory for the JSON files. Each package will have its own JSON file.")
	cmd.MarkFlagRequired("out-dir")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		root := docDumpOpts.ProjectBaseDir
		docs, err := parseProject(root)
		if err != nil {
			log.Fatalf("failed to parse project: %v", err)
		}
		// Create the output directory if it doesn't exist
		if err := os.MkdirAll(docDumpOpts.OutDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
		for _, apiDoc := range docs {
			filename := filepath.Join(docDumpOpts.OutDir, fmt.Sprintf("%s.json", apiDoc.Package))
			f, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", filename, err)
			}
			enc := json.NewEncoder(f)
			defer f.Close()
			enc.SetIndent("", "  ")
			if err := enc.Encode(apiDoc); err != nil {
				return fmt.Errorf("failed to encode json for package %s: %v", apiDoc.Package, err)
			}
		}
		cmd.Printf("API docs dumped to %s\n", docDumpOpts.OutDir)
		return nil
	}
	return cmd
}

func parseProject(root string) ([]APIDocs, error) {
	var result []APIDocs
	var parsed []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor and hidden directories to avoid infinite loops
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			if path == root {
				return nil
			}
			if contains(parsed, path) {
				return nil
			}
			pkgs, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.ParseComments)
			if err != nil {
				return nil // skip dirs without Go files
			}
			parsed = append(parsed, path)
			for _, pkg := range pkgs {
				docPkg := godoc.New(pkg, path, godoc.AllDecls)
				api := APIDocs{
					Package:   docPkg.Name,
					Functions: []FuncDoc{},
					Types:     []TypeDoc{},
					Consts:    []ValueDoc{},
					Vars:      []ValueDoc{},
				}
				// functions
				for _, f := range docPkg.Funcs {
					// Build function signature: "func Name(params) (results)"
					signature := constructSignature(f)
					api.Functions = append(api.Functions, FuncDoc{
						Name:      f.Name,
						Signature: signature,
						Doc:       strings.TrimSpace(f.Doc),
					})
				}

				// types + methods
				for _, t := range docPkg.Types {
					td := TypeDoc{
						Name:    t.Name,
						Doc:     strings.TrimSpace(t.Doc),
						Methods: []MethodDoc{},
					}
					for _, m := range t.Methods {
						signature := constructSignature(m)
						td.Methods = append(td.Methods, MethodDoc{
							Name:      m.Name,
							Signature: signature,
							Doc:       strings.TrimSpace(m.Doc),
						})
					}
					api.Types = append(api.Types, td)
				}

				// consts
				for _, c := range docPkg.Consts {
					api.Consts = append(api.Consts, ValueDoc{
						Names: c.Names,
						Doc:   strings.TrimSpace(c.Doc),
					})
				}

				// vars
				for _, v := range docPkg.Vars {
					api.Vars = append(api.Vars, ValueDoc{
						Names: v.Names,
						Doc:   strings.TrimSpace(v.Doc),
					})
				}

				result = append(result, api)
			}
		}
		return nil
	})
	return result, err
}

func constructSignature(f *godoc.Func) string {
	signature := ""
	if f.Decl != nil {
		// Use go/printer to print the function declaration
		var buf strings.Builder
		buf.WriteString("func ")
		buf.WriteString(f.Name)
		buf.WriteString("(")
		// Parameters
		for i, param := range f.Decl.Type.Params.List {
			// param.Names can be empty for unnamed parameters
			names := []string{}
			for _, n := range param.Names {
				names = append(names, n.Name)
			}
			paramName := strings.Join(names, ", ")
			if paramName != "" {
				buf.WriteString(paramName)
				buf.WriteString(" ")
			}
			// Print type
			buf.WriteString(exprString(param.Type))
			if i < len(f.Decl.Type.Params.List)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(")")
		// Results
		if f.Decl.Type.Results != nil && len(f.Decl.Type.Results.List) > 0 {
			buf.WriteString(" ")
			if len(f.Decl.Type.Results.List) > 1 {
				buf.WriteString("(")
			}
			for i, result := range f.Decl.Type.Results.List {
				names := []string{}
				for _, n := range result.Names {
					names = append(names, n.Name)
				}
				resultName := strings.Join(names, ", ")
				if resultName != "" {
					buf.WriteString(resultName)
					buf.WriteString(" ")
				}
				buf.WriteString(exprString(result.Type))
				if i < len(f.Decl.Type.Results.List)-1 {
					buf.WriteString(", ")
				}
			}
			if len(f.Decl.Type.Results.List) > 1 {
				buf.WriteString(")")
			}
		}
		signature = buf.String()
	}
	return signature
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// strip extra spaces & newlines from decl text
func cleanDecl(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// exprString converts an ast.Expr to its string representation.
func exprString(expr interface{}) string {
	switch e := expr.(type) {
	case *godoc.Value:
		return strings.Join(e.Names, ", ")
	case *godoc.Type:
		return e.Name
	default:
		// fallback for ast.Expr
		return fmt.Sprintf("%v", expr)
	}
}
