package registry

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Tool represents a known AI coding agent/tool and its configuration paths.
type Tool struct {
	// Name is the display name of the tool.
	Name string

	// Description is a short description of the tool.
	Description string

	// GlobalConfigPath is the user-level config file the tool reads.
	// Empty if the tool has no global config. Supports ~ prefix.
	GlobalConfigPath string

	// RepoFileName is the filename the tool looks for at repository root.
	// Empty if the tool does not read repo-level files.
	RepoFileName string

	// ReadsAgentsMD indicates whether the tool natively reads AGENTS.md.
	ReadsAgentsMD bool

	// DetectPaths are directories or files whose existence indicates the tool
	// is installed. Checked in order; first match wins. Supports ~ prefix.
	DetectPaths []string

	// DetectCommands are CLI commands checked via PATH lookup.
	// Checked only if DetectPaths yields no match.
	DetectCommands []string
}

// Detected holds the result of a tool detection check.
type Detected struct {
	Tool    Tool
	Method  string // "path", "command", or ""
	Details string // which path or command matched
}

// All returns the complete registry of known AI coding tools.
// Add new tools here. Keep alphabetical by Name.
func All() []Tool {
	return []Tool{
		{
			Name:             "Aider",
			Description:      "AI pair programming in your terminal",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{},
			DetectCommands:   []string{"aider"},
		},
		{
			Name:             "Amp",
			Description:      "AI-native code editor",
			GlobalConfigPath: "~/.config/AGENTS.md",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.config/amp"},
			DetectCommands:   []string{"amp"},
		},
		{
			Name:             "Antigravity",
			Description:      "Google cloud IDE with Gemini integration",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{},
			DetectCommands:   []string{},
		},
		{
			Name:             "Autohand",
			Description:      "AI coding assistant",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.config/autohand"},
			DetectCommands:   []string{"autohand"},
		},
		{
			Name:             "Claude Code",
			Description:      "Anthropic CLI for agentic coding",
			GlobalConfigPath: "~/.claude/CLAUDE.md",
			RepoFileName:     "CLAUDE.md",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.claude"},
			DetectCommands:   []string{"claude"},
		},
		{
			Name:             "Cline",
			Description:      "Autonomous coding agent for VS Code",
			GlobalConfigPath: "",
			RepoFileName:     "",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/Documents/Cline"},
			DetectCommands:   []string{},
		},
		{
			Name:             "Continue",
			Description:      "Open-source AI code assistant",
			GlobalConfigPath: "~/.continue/config.yaml",
			RepoFileName:     "",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.continue"},
			DetectCommands:   []string{"continue"},
		},
		{
			Name:             "Cursor",
			Description:      "AI-first code editor",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.cursor"},
			DetectCommands:   []string{"cursor"},
		},
		{
			Name:             "Factory (Droid)",
			Description:      "AI software engineering platform",
			GlobalConfigPath: "~/.factory/AGENTS.md",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.factory"},
			DetectCommands:   []string{"factory"},
		},
		{
			Name:             "Gemini CLI",
			Description:      "Google Gemini command-line tool",
			GlobalConfigPath: "~/.gemini/GEMINI.md",
			RepoFileName:     "GEMINI.md",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.gemini"},
			DetectCommands:   []string{"gemini"},
		},
		{
			Name:             "GitHub Copilot",
			Description:      "AI pair programmer by GitHub",
			GlobalConfigPath: "",
			RepoFileName:     ".github/copilot-instructions.md",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{},
			DetectCommands:   []string{"gh"},
		},
		{
			Name:             "Goose",
			Description:      "AI developer agent by Block",
			GlobalConfigPath: "~/.config/goose/.goosehints",
			RepoFileName:     ".goosehints",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.config/goose"},
			DetectCommands:   []string{"goose"},
		},
		{
			Name:             "Junie",
			Description:      "JetBrains AI coding agent",
			GlobalConfigPath: "",
			RepoFileName:     ".junie/AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.junie"},
			DetectCommands:   []string{},
		},
		{
			Name:             "Kilo Code",
			Description:      "AI coding assistant",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.kilo"},
			DetectCommands:   []string{"kilo"},
		},
		{
			Name:             "Codex CLI",
			Description:      "OpenAI command-line coding agent",
			GlobalConfigPath: "~/.codex/AGENTS.md",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.codex"},
			DetectCommands:   []string{"codex"},
		},
		{
			Name:             "OpenCode",
			Description:      "Terminal-based AI coding assistant",
			GlobalConfigPath: "~/.config/opencode/AGENTS.md",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.config/opencode"},
			DetectCommands:   []string{"opencode"},
		},
		{
			Name:             "RooCode",
			Description:      "AI coding assistant for VS Code",
			GlobalConfigPath: "",
			RepoFileName:     ".roo/rules/rules.md",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.roo"},
			DetectCommands:   []string{},
		},
		{
			Name:             "Windsurf",
			Description:      "AI-powered IDE by Codeium",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.codeium"},
			DetectCommands:   []string{"windsurf"},
		},
		{
			Name:             "Zed",
			Description:      "High-performance multiplayer code editor",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.config/zed"},
			DetectCommands:   []string{"zed"},
		},
	}
}

// DetectAll checks which tools from the registry are installed on this system.
func DetectAll() []Detected {
	var results []Detected
	for _, tool := range All() {
		if d := detectTool(tool); d != nil {
			results = append(results, *d)
		}
	}
	return results
}

// DetectInstalled returns only tools that were found on the system.
func DetectInstalled() []Detected {
	return DetectAll()
}

// ToolsWithGlobalConfig returns tools that have a user-level config path.
func ToolsWithGlobalConfig() []Tool {
	var tools []Tool
	for _, t := range All() {
		if t.GlobalConfigPath != "" {
			tools = append(tools, t)
		}
	}
	return tools
}

// ToolsReadingAgentsMD returns tools that natively read AGENTS.md.
func ToolsReadingAgentsMD() []Tool {
	var tools []Tool
	for _, t := range All() {
		if t.ReadsAgentsMD {
			tools = append(tools, t)
		}
	}
	return tools
}

func detectTool(tool Tool) *Detected {
	homeDir, _ := os.UserHomeDir()

	// Check paths first
	for _, p := range tool.DetectPaths {
		expanded := expandHome(p, homeDir)
		if _, err := os.Stat(expanded); err == nil {
			return &Detected{
				Tool:    tool,
				Method:  "path",
				Details: expanded,
			}
		}
	}

	// Check commands
	for _, cmd := range tool.DetectCommands {
		if path, err := exec.LookPath(cmd); err == nil {
			return &Detected{
				Tool:    tool,
				Method:  "command",
				Details: path,
			}
		}
	}

	return nil
}

func expandHome(path, homeDir string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}
