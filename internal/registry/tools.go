package registry

import (
	"os"
	"os/exec"
	"path/filepath"
)

// AgentsMDIntegration describes the least invasive supported way for a tool
// to consume AGENTS.md. It is advisory metadata: Agentlink still creates only
// symlinks and never writes tool-specific import or settings files.
type AgentsMDIntegration string

const (
	IntegrationNative       AgentsMDIntegration = "native"
	IntegrationConfigurable AgentsMDIntegration = "configurable"
	IntegrationImport       AgentsMDIntegration = "import"
	IntegrationSymlink      AgentsMDIntegration = "symlink"
	IntegrationUnsupported  AgentsMDIntegration = "unsupported"
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

	// PreferredIntegration overrides the integration inferred from
	// ReadsAgentsMD and RepoFileName. Use it when a tool supports a less
	// invasive option such as configuration or an explicit import.
	PreferredIntegration AgentsMDIntegration

	// SupportsNestedRepoFile indicates that the tool discovers its repo file
	// below the repository root. Leave false unless the behavior is documented
	// and deterministic; nested scanning fails closed for unknown tools.
	SupportsNestedRepoFile bool

	// NestedSupportReference identifies the public documentation used to
	// justify SupportsNestedRepoFile. Keep this empty when nested discovery is
	// not enabled so registry validation can fail closed on unsupported claims.
	NestedSupportReference string

	// DetectPaths are directories or files whose existence indicates the tool
	// is installed. Checked in order; first match wins. Supports ~ prefix.
	DetectPaths []string

	// DetectCommands are CLI commands checked via PATH lookup.
	// Checked only if DetectPaths yields no match.
	DetectCommands []string
}

// AgentsMDIntegration returns the preferred AGENTS.md integration for a tool.
func (t Tool) AgentsMDIntegration() AgentsMDIntegration {
	if t.PreferredIntegration != "" {
		return t.PreferredIntegration
	}
	if t.ReadsAgentsMD {
		return IntegrationNative
	}
	if t.RepoFileName != "" {
		return IntegrationSymlink
	}
	return IntegrationUnsupported
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
			Name:                   "Claude Code",
			Description:            "Anthropic CLI for agentic coding",
			GlobalConfigPath:       "~/.claude/CLAUDE.md",
			RepoFileName:           "CLAUDE.md",
			ReadsAgentsMD:          false,
			PreferredIntegration:   IntegrationImport,
			SupportsNestedRepoFile: true,
			NestedSupportReference: "https://code.claude.com/docs/en/memory",
			DetectPaths:            []string{"~/.claude"},
			DetectCommands:         []string{"claude"},
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
			Name:             "Crush",
			Description:      "Charm terminal coding agent",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.config/crush"},
			DetectCommands:   []string{"crush"},
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
			Name:                   "Gemini CLI",
			Description:            "Google Gemini command-line tool",
			GlobalConfigPath:       "~/.gemini/GEMINI.md",
			RepoFileName:           "GEMINI.md",
			ReadsAgentsMD:          false,
			PreferredIntegration:   IntegrationConfigurable,
			SupportsNestedRepoFile: true,
			NestedSupportReference: "https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/gemini-md.md",
			DetectPaths:            []string{"~/.gemini"},
			DetectCommands:         []string{"gemini"},
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
			Name:             "Kiro",
			Description:      "AWS agentic IDE",
			GlobalConfigPath: "",
			RepoFileName:     "AGENTS.md",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.kiro"},
			DetectCommands:   []string{"kiro"},
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
			Name:             "OpenClaw",
			Description:      "Personal AI assistant with an agent workspace",
			GlobalConfigPath: "~/.openclaw/workspace/AGENTS.md",
			RepoFileName:     "",
			ReadsAgentsMD:    true,
			DetectPaths:      []string{"~/.openclaw"},
			DetectCommands:   []string{"openclaw"},
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
			Name:             "Qwen Code",
			Description:      "Alibaba open-source CLI for agentic coding",
			GlobalConfigPath: "~/.qwen/QWEN.md",
			RepoFileName:     "QWEN.md",
			ReadsAgentsMD:    false,
			DetectPaths:      []string{"~/.qwen"},
			DetectCommands:   []string{"qwen"},
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
