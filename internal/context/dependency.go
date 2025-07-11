package context

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DependencyAnalyzer provides dependency graph construction for projects
type DependencyAnalyzer interface {
	// AnalyzeDependencies builds a dependency graph for the given files
	AnalyzeDependencies(ctx context.Context, files []FileInfo) (*DependencyGraph, error)
	
	// GetFileDependencies returns direct dependencies for a single file
	GetFileDependencies(ctx context.Context, filePath string) ([]string, error)
	
	// GetDependents returns files that depend on the given file
	GetDependents(graph *DependencyGraph, filePath string) []string
	
	// CalculateCentrality calculates importance of a file in the dependency graph
	CalculateCentrality(graph *DependencyGraph, filePath string) float64
}

// GoDependencyAnalyzer analyzes Go project dependencies
type GoDependencyAnalyzer struct {
	projectRoot string
	moduleInfo  *GoModuleInfo
}

// GoModuleInfo contains Go module information
type GoModuleInfo struct {
	ModulePath string
	GoVersion  string
}

// NewGoDependencyAnalyzer creates a new Go dependency analyzer
func NewGoDependencyAnalyzer(projectRoot string) *GoDependencyAnalyzer {
	analyzer := &GoDependencyAnalyzer{
		projectRoot: projectRoot,
	}
	
	// Try to load module info
	analyzer.moduleInfo = analyzer.loadModuleInfo()
	
	return analyzer
}

// AnalyzeDependencies builds a complete dependency graph
func (a *GoDependencyAnalyzer) AnalyzeDependencies(ctx context.Context, files []FileInfo) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: []DependencyEdge{},
	}
	
	// First pass: Create nodes for all Go files
	goFiles := []FileInfo{}
	for _, file := range files {
		if file.Language == "go" && !strings.Contains(file.Path, "_test.go") {
			relPath, _ := filepath.Rel(a.projectRoot, file.Path)
			graph.Nodes[relPath] = &DependencyNode{
				Path:         relPath,
				Imports:      []string{},
				Exports:      []string{},
				Dependencies: []string{},
				Dependents:   []string{},
			}
			goFiles = append(goFiles, file)
		}
	}
	
	// Second pass: Analyze imports and build edges
	for _, file := range goFiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		imports, exports, err := a.analyzeGoFile(file.Path)
		if err != nil {
			continue // Skip files with parse errors
		}
		
		relPath, _ := filepath.Rel(a.projectRoot, file.Path)
		node := graph.Nodes[relPath]
		node.Imports = imports
		node.Exports = exports
		
		// Map imports to local files
		for _, imp := range imports {
			if depFile := a.resolveImportToFile(imp, goFiles); depFile != "" {
				depRelPath, _ := filepath.Rel(a.projectRoot, depFile)
				
				// Update dependency relationships
				node.Dependencies = append(node.Dependencies, depRelPath)
				if depNode, exists := graph.Nodes[depRelPath]; exists {
					depNode.Dependents = append(depNode.Dependents, relPath)
				}
				
				// Create edge
				graph.Edges = append(graph.Edges, DependencyEdge{
					From:     relPath,
					To:       depRelPath,
					Type:     "import",
					Strength: 1.0, // Direct import has full strength
				})
			}
		}
	}
	
	return graph, nil
}

// GetFileDependencies returns direct dependencies for a single file
func (a *GoDependencyAnalyzer) GetFileDependencies(ctx context.Context, filePath string) ([]string, error) {
	imports, _, err := a.analyzeGoFile(filePath)
	if err != nil {
		return nil, err
	}
	
	// Filter to only local project imports
	localDeps := []string{}
	for _, imp := range imports {
		if a.isLocalImport(imp) {
			localDeps = append(localDeps, imp)
		}
	}
	
	return localDeps, nil
}

// GetDependents returns files that depend on the given file
func (a *GoDependencyAnalyzer) GetDependents(graph *DependencyGraph, filePath string) []string {
	relPath, _ := filepath.Rel(a.projectRoot, filePath)
	
	if node, exists := graph.Nodes[relPath]; exists {
		return node.Dependents
	}
	
	return []string{}
}

// CalculateCentrality calculates the importance of a file in the dependency graph
func (a *GoDependencyAnalyzer) CalculateCentrality(graph *DependencyGraph, filePath string) float64 {
	relPath, _ := filepath.Rel(a.projectRoot, filePath)
	
	node, exists := graph.Nodes[relPath]
	if !exists {
		return 0.0
	}
	
	// Simple centrality: combination of in-degree and out-degree
	// Normalized by total number of nodes
	totalNodes := len(graph.Nodes)
	if totalNodes <= 1 {
		return 0.5
	}
	
	inDegree := float64(len(node.Dependents))
	outDegree := float64(len(node.Dependencies))
	
	// Files that are depended upon by many others are important
	// Files that depend on many others are also somewhat important (integration points)
	centrality := (inDegree*2 + outDegree) / float64(3*(totalNodes-1))
	
	return min(1.0, centrality)
}

// analyzeGoFile parses a Go file and extracts imports and exports
func (a *GoDependencyAnalyzer) analyzeGoFile(filePath string) (imports []string, exports []string, err error) {
	fset := token.NewFileSet()
	
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	
	file, err := parser.ParseFile(fset, filePath, src, parser.ImportsOnly)
	if err != nil {
		return nil, nil, err
	}
	
	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}
	
	// For exports, we need a full parse
	file, err = parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		return imports, nil, err // Return imports even if full parse fails
	}
	
	// Extract exported types, functions, and variables
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				exports = append(exports, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						exports = append(exports, s.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							exports = append(exports, name.Name)
						}
					}
				}
			}
		}
	}
	
	return imports, exports, nil
}

// resolveImportToFile maps an import path to a local file path
func (a *GoDependencyAnalyzer) resolveImportToFile(importPath string, files []FileInfo) string {
	// Skip standard library and external imports
	if !a.isLocalImport(importPath) {
		return ""
	}
	
	// Convert import path to potential file paths
	var searchPath string
	if a.moduleInfo != nil && strings.HasPrefix(importPath, a.moduleInfo.ModulePath) {
		// Module-relative import
		relPath := strings.TrimPrefix(importPath, a.moduleInfo.ModulePath)
		relPath = strings.TrimPrefix(relPath, "/")
		searchPath = filepath.Join(a.projectRoot, relPath)
	} else {
		// Try as relative path
		searchPath = filepath.Join(a.projectRoot, importPath)
	}
	
	// Look for matching files
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if strings.HasPrefix(dir, searchPath) {
			// Check if this is the main file of the package
			if filepath.Base(file.Path) == "doc.go" || 
			   strings.HasSuffix(file.Path, "_test.go") {
				continue
			}
			return file.Path
		}
	}
	
	return ""
}

// isLocalImport checks if an import is from the local project
func (a *GoDependencyAnalyzer) isLocalImport(importPath string) bool {
	// Standard library imports don't contain dots (except for vendored)
	if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/vendor/") {
		return false
	}
	
	// Check if it's our module
	if a.moduleInfo != nil {
		return strings.HasPrefix(importPath, a.moduleInfo.ModulePath)
	}
	
	// Check common patterns
	return !strings.HasPrefix(importPath, "github.com/") &&
		!strings.HasPrefix(importPath, "golang.org/") &&
		!strings.HasPrefix(importPath, "google.golang.org/") &&
		!strings.HasPrefix(importPath, "gopkg.in/")
}

// loadModuleInfo loads Go module information from go.mod
func (a *GoDependencyAnalyzer) loadModuleInfo() *GoModuleInfo {
	goModPath := filepath.Join(a.projectRoot, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return nil
	}
	defer file.Close()
	
	info := &GoModuleInfo{}
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "module ") {
			info.ModulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		} else if strings.HasPrefix(line, "go ") {
			info.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
		
		// Stop after finding both
		if info.ModulePath != "" && info.GoVersion != "" {
			break
		}
	}
	
	if info.ModulePath == "" {
		return nil
	}
	
	return info
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MultilanguageDependencyAnalyzer can analyze dependencies for multiple languages
type MultilanguageDependencyAnalyzer struct {
	analyzers map[string]DependencyAnalyzer
}

// NewMultilanguageDependencyAnalyzer creates a dependency analyzer that supports multiple languages
func NewMultilanguageDependencyAnalyzer(projectRoot string) *MultilanguageDependencyAnalyzer {
	return &MultilanguageDependencyAnalyzer{
		analyzers: map[string]DependencyAnalyzer{
			"go": NewGoDependencyAnalyzer(projectRoot),
			// Can add more language analyzers here
		},
	}
}

// AnalyzeDependencies delegates to appropriate language analyzer
func (m *MultilanguageDependencyAnalyzer) AnalyzeDependencies(ctx context.Context, files []FileInfo) (*DependencyGraph, error) {
	// Group files by language
	filesByLang := make(map[string][]FileInfo)
	for _, file := range files {
		filesByLang[file.Language] = append(filesByLang[file.Language], file)
	}
	
	// Use the analyzer for the dominant language
	// In the future, could merge graphs from multiple analyzers
	var dominantLang string
	maxFiles := 0
	for lang, langFiles := range filesByLang {
		if len(langFiles) > maxFiles {
			maxFiles = len(langFiles)
			dominantLang = lang
		}
	}
	
	if analyzer, exists := m.analyzers[dominantLang]; exists {
		return analyzer.AnalyzeDependencies(ctx, filesByLang[dominantLang])
	}
	
	// Return empty graph if no suitable analyzer
	return &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: []DependencyEdge{},
	}, nil
}

// GetFileDependencies is not implemented for multilanguage analyzer
func (m *MultilanguageDependencyAnalyzer) GetFileDependencies(ctx context.Context, filePath string) ([]string, error) {
	return nil, fmt.Errorf("not implemented for multilanguage analyzer")
}

// GetDependents returns files that depend on the given file
func (m *MultilanguageDependencyAnalyzer) GetDependents(graph *DependencyGraph, filePath string) []string {
	relPath, _ := filepath.Rel(".", filePath)
	
	if node, exists := graph.Nodes[relPath]; exists {
		return node.Dependents
	}
	
	return []string{}
}

// CalculateCentrality delegates to appropriate analyzer
func (m *MultilanguageDependencyAnalyzer) CalculateCentrality(graph *DependencyGraph, filePath string) float64 {
	// Use generic centrality calculation
	relPath, _ := filepath.Rel(".", filePath)
	
	node, exists := graph.Nodes[relPath]
	if !exists {
		return 0.0
	}
	
	totalNodes := len(graph.Nodes)
	if totalNodes <= 1 {
		return 0.5
	}
	
	inDegree := float64(len(node.Dependents))
	outDegree := float64(len(node.Dependencies))
	
	centrality := (inDegree*2 + outDegree) / float64(3*(totalNodes-1))
	
	return min(1.0, centrality)
}