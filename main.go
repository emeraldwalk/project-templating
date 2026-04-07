package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

func main() {
	configPath := flag.String("config", "", "Path to a JSON config file for extra variables")
	templateArg := flag.String("template", "", "Template directory name or path to process")
	templateRoot := flag.String("template-root", "", "Override the root directory searched for templates (default: <project>/templates)")
	destDir := flag.String("dest", ".", "Destination directory for generated files")
	flag.Parse()

	// Resolve --template: check templates/<arg> first, then relative to cwd
	srcDir := resolveTemplateDir(*templateArg, *templateRoot)

	// 1. Fail-fast: Ensure we are in a Git repo
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Current directory is not a git repository.\n")
		os.Exit(1)
	}

	// 2. Initialize the Flat Context Map
	ctx := make(map[string]any)

	// 3. Add Built-in "Special" Variables
	cwd, _ := os.Getwd()
	isWorktree, relSource, absTarget, mainFolderPath, mainFolderBasename := getGitMountInfo()

	bgColor := "#" + GenerateColorFromPath(cwd)
	ctx["BG_COLOR"] = bgColor
	ctx["FG_COLOR"] = GetContrastingForeground(bgColor)
	ctx["IS_GIT_WORKTREE"] = isWorktree
	ctx["GIT_REL_SOURCE"] = relSource
	ctx["GIT_ABS_TARGET"] = absTarget
	ctx["GIT_BRANCH"] = getGitBranch()
	ctx["GIT_WORKTREE_MAIN_FOLDER_PATH"] = mainFolderPath
	ctx["GIT_WORKTREE_MAIN_FOLDER_BASENAME"] = mainFolderBasename

	// 4. Merge JSON Config File (if provided)
	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err == nil {
			json.Unmarshal(data, &ctx)
		}
	}

	// 5. Merge CLI trailing args (e.g., APP_NAME=my-service)
	for _, arg := range flag.Args() {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			ctx[parts[0]] = parts[1]
		}
	}

	// 6. Process Template Folder
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Calculate output path
		relPath, _ := filepath.Rel(srcDir, path)
		targetPath := filepath.Join(*destDir, relPath)
		os.MkdirAll(filepath.Dir(targetPath), 0755)

		// Parse and Execute
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		f, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer f.Close()

		fmt.Printf("Generating: %s\n", targetPath)
		return tmpl.Execute(f, ctx)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Template Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveTemplateDir finds the template directory for the given argument.
// The project root is derived from the binary location: the binary lives in
// <project>/bin/, so the project root is its parent directory.
// If arg is empty, defaults to <project>/templates/.
// If a relative path is given, it first checks <project>/templates/<arg>,
// then falls back to arg relative to cwd. Absolute paths are used as-is.
func resolveTemplateDir(arg, templateRoot string) string {
	root := templateRoot
	if root == "" {
		if exe, err := os.Executable(); err == nil {
			root = filepath.Join(filepath.Dir(filepath.Dir(exe)), "templates")
		}
	}

	if arg == "" {
		if root != "" {
			return root
		}
		return "templates"
	}
	if filepath.IsAbs(arg) {
		return arg
	}
	if root != "" {
		candidate := filepath.Join(root, arg)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return arg
}

// Git Logic
func getGitMountInfo() (bool, string, string, string, string) {
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return false, "", "", "", ""
	}

	absTarget, _ := filepath.Abs(strings.TrimSpace(string(out)))
	cwd, _ := os.Getwd()
	relSource, _ := filepath.Rel(cwd, absTarget)

	// If common dir is just ".git", we are in primary tree
	isWorktree := !strings.HasSuffix(filepath.ToSlash(absTarget), "/.git") &&
		filepath.Base(absTarget) != ".git"

	// Find the .git dir within absTarget, then go up one level to get the main worktree folder.
	// Primary tree: absTarget = /repo/.git         → main folder = /repo
	// Worktree:     absTarget = /repo/.git/worktrees/foo → main folder = /repo
	gitDirIndex := strings.Index(filepath.ToSlash(absTarget), "/.git")
	mainFolderPath := absTarget
	if gitDirIndex >= 0 {
		mainFolderPath = absTarget[:gitDirIndex]
	}
	mainFolderBasename := filepath.Base(mainFolderPath)

	return isWorktree, relSource, absTarget, mainFolderPath, mainFolderBasename
}

func getGitBranch() string {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "unknown"
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		// Fallback for detached HEAD
		out, _ = exec.Command("git", "rev-parse", "--short", "HEAD").Output()
		return strings.TrimSpace(string(out))
	}
	return branch
}

// Color Logic
func GenerateColorFromPath(path string) string {
	hash := md5.Sum([]byte(path))
	r := hash[0] ^ hash[8]
	g := hash[5] ^ hash[13]
	b := hash[10] ^ hash[2]
	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

func GetContrastingForeground(bgColor string) string {
	bgColor = strings.TrimPrefix(bgColor, "#")
	if len(bgColor) != 6 {
		return "#ffffff"
	}
	r, _ := strconv.ParseInt(bgColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(bgColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(bgColor[4:6], 16, 64)

	luminance := (0.2126 * float64(r)) + (0.7152 * float64(g)) + (0.0722 * float64(b))
	if luminance > 128 {
		return "#000000"
	}
	return "#ffffff"
}
