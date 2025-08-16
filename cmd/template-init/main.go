package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

type Config struct {
	OldModule   string
	NewModule   string
	ProjectName string
	DryRun      bool
}

func main() {
	// Auto-detect current module from go.mod
	oldModule, err := detectCurrentModule()
	if err != nil {
		fmt.Printf("%sError detecting current module: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	fmt.Printf("%s🚀 Template Initialization%s\n", colorBlue, colorReset)
	fmt.Printf("Current module detected: %s%s%s\n\n", colorYellow, oldModule, colorReset)

	// Get new module from user
	fmt.Print("Enter new repository module (e.g., github.com/yourorg/project): ")
	reader := bufio.NewReader(os.Stdin)
	newModule, _ := reader.ReadString('\n')
	newModule = strings.TrimSpace(newModule)

	if newModule == "" {
		fmt.Printf("%sError: New module name is required%s\n", colorRed, colorReset)
		os.Exit(1)
	}

	// Validate module format
	moduleRegex := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !moduleRegex.MatchString(newModule) {
		fmt.Printf("%sError: Invalid module name format%s\n", colorRed, colorReset)
		os.Exit(1)
	}

	// Extract project name from module
	parts := strings.Split(newModule, "/")
	defaultProjectName := parts[len(parts)-1]

	fmt.Printf("Enter project name (default: %s): ", defaultProjectName)
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		projectName = defaultProjectName
	}

	config := Config{
		OldModule:   oldModule,
		NewModule:   newModule,
		ProjectName: projectName,
		DryRun:      false,
	}

	fmt.Printf("\n%s📋 Configuration:%s\n", colorBlue, colorReset)
	fmt.Printf("  Old module: %s%s%s\n", colorYellow, config.OldModule, colorReset)
	fmt.Printf("  New module: %s%s%s\n", colorGreen, config.NewModule, colorReset)
	fmt.Printf("  Project name: %s%s%s\n", colorGreen, config.ProjectName, colorReset)
	fmt.Println()

	fmt.Print("Continue with template initialization? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("❌ Template initialization cancelled")
		return
	}

	if err := processTemplate(config); err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Automatically set git remote if possible
	if err := setGitRemote(config.NewModule); err != nil {
		fmt.Printf("%sWarning: Could not set git remote automatically: %v%s\n", colorYellow, err, colorReset)
		fmt.Printf("Please set manually: git remote set-url origin %s.git\n", config.NewModule)
	} else {
		fmt.Printf("%s✅ Git remote updated to: %s.git%s\n", colorGreen, config.NewModule, colorReset)
	}

	fmt.Printf("\n%s✅ Template initialization completed!%s\n", colorGreen, colorReset)
	showNextSteps(config.NewModule)
}

func detectCurrentModule() (string, error) {
	goModPath := "go.mod"
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return "", fmt.Errorf("go.mod not found in current directory")
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return module, nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

func processTemplate(config Config) error {
	fmt.Printf("%s📝 Processing files...%s\n", colorBlue, colorReset)

	// File patterns and their replacement rules
	patterns := []FilePattern{
		{
			Pattern:     "go.mod",
			Description: "Go module file",
			Replacements: []Replacement{
				{Old: config.OldModule, New: config.NewModule},
			},
		},
		{
			Pattern:     "**/*.go",
			Description: "Go source files",
			Replacements: []Replacement{
				{Old: config.OldModule, New: config.NewModule},
			},
		},
		{
			Pattern:     "**/*.proto",
			Description: "Protocol buffer files",
			Replacements: []Replacement{
				{Old: config.OldModule, New: config.NewModule},
			},
		},
		{
			Pattern:     "buf.yaml",
			Description: "Buf configuration",
			Replacements: []Replacement{
				{Old: config.OldModule, New: config.NewModule},
			},
		},
		{
			Pattern:     "files/**/*.{yaml,yml,json}",
			Description: "Configuration files",
			Replacements: []Replacement{
				{Old: extractProjectName(config.OldModule), New: config.ProjectName},
				{Old: strings.ReplaceAll(extractProjectName(config.OldModule), "-", "_"), New: strings.ReplaceAll(config.ProjectName, "-", "_")},
			},
		},
		{
			Pattern:     "docker-compose.yml",
			Description: "Docker Compose configuration",
			Replacements: []Replacement{
				{Old: strings.ReplaceAll(extractProjectName(config.OldModule), "-", "_"), New: strings.ReplaceAll(config.ProjectName, "-", "_")},
			},
		},
		{
			Pattern:     "atlas.hcl",
			Description: "Atlas migration configuration",
			Replacements: []Replacement{
				{Old: strings.ReplaceAll(extractProjectName(config.OldModule), "-", "_"), New: strings.ReplaceAll(config.ProjectName, "-", "_")},
			},
		},
	}

	for _, pattern := range patterns {
		if err := processPattern(pattern, config.DryRun); err != nil {
			return fmt.Errorf("processing %s: %w", pattern.Description, err)
		}
	}

	return nil
}

func extractProjectName(module string) string {
	parts := strings.Split(module, "/")
	return parts[len(parts)-1]
}

type FilePattern struct {
	Pattern      string
	Description  string
	Replacements []Replacement
}

type Replacement struct {
	Old string
	New string
}

func processPattern(pattern FilePattern, dryRun bool) error {
	files, err := findFiles(pattern.Pattern)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	fmt.Printf("  → %s\n", pattern.Description)

	for _, file := range files {
		changed, err := processFile(file, pattern.Replacements, dryRun)
		if err != nil {
			return fmt.Errorf("processing %s: %w", file, err)
		}
		if changed {
			fmt.Printf("    - %s\n", file)
		}
	}

	return nil
}

func findFiles(pattern string) ([]string, error) {
	var files []string

	// Handle special patterns
	if pattern == "**/*.go" {
		return findByExtension(".go"), nil
	}
	if pattern == "**/*.proto" {
		return findByExtension(".proto"), nil
	}
	if pattern == "files/**/*.{yaml,yml,json}" {
		return findConfigFiles(), nil
	}

	// Single file
	if _, err := os.Stat(pattern); err == nil {
		files = append(files, pattern)
	}

	return files, nil
}

func findByExtension(ext string) []string {
	var files []string

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and hidden files
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Skip vendor, node_modules, and bin directories
		if strings.Contains(path, "vendor/") ||
			strings.Contains(path, "node_modules/") ||
			strings.Contains(path, "bin/") {
			return nil
		}

		if strings.HasSuffix(path, ext) {
			files = append(files, path)
		}

		return nil
	})

	return files
}

func findConfigFiles() []string {
	var files []string

	if _, err := os.Stat("files"); err != nil {
		return files
	}

	filepath.WalkDir("files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			files = append(files, path)
		}

		return nil
	})

	return files
}

func processFile(filename string, replacements []Replacement, dryRun bool) (bool, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	originalContent := string(content)
	newContent := originalContent

	// Apply all replacements
	for _, repl := range replacements {
		if strings.Contains(newContent, repl.Old) {
			newContent = strings.ReplaceAll(newContent, repl.Old, repl.New)
		}
	}

	// Check if file was changed
	if newContent == originalContent {
		return false, nil
	}

	// Write file if not dry run
	if !dryRun {
		if err := os.WriteFile(filename, []byte(newContent), 0644); err != nil {
			return false, err
		}
	}

	return true, nil
}

func setGitRemote(newModule string) error {
	// Check if git is available and this is a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository")
	}

	// Construct the git URL from the module name
	gitURL := fmt.Sprintf("https://%s.git", newModule)
	
	// Execute git remote set-url origin command
	cmd := fmt.Sprintf("git remote set-url origin %s", gitURL)
	if err := executeCommand(cmd); err != nil {
		return fmt.Errorf("failed to set git remote: %w", err)
	}

	return nil
}

func executeCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}

	return nil
}

func showNextSteps(newModule string) {
	fmt.Printf("\n%s🔗 Makefile will automatically run:%s\n", colorBlue, colorReset)
	fmt.Println("  1. go mod tidy")
	fmt.Println("  2. make generate")
	fmt.Println("  3. make test")
	fmt.Printf("\n%s🔗 Manual steps:%s\n", colorBlue, colorReset)
	fmt.Printf("  1. Review git remote: %s.git\n", newModule)
	fmt.Println("  2. Update README.md with your project details")
	fmt.Println("  3. Start development: make dev")
	fmt.Printf("\n%s🎉 Happy coding!%s\n", colorGreen, colorReset)
}
