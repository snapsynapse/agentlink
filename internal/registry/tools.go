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
	// IntegrationReference and ReviewedOn record the vendor evidence for the
	// advertised instruction path. Notes distinguish client/version caveats.
	IntegrationReference string
	ReviewedOn           string
	IntegrationNotes     string

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
			Name:                 "Aider",
			Description:          "AI pair programming in your terminal",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        false,
			DetectPaths:          []string{},
			DetectCommands:       []string{"aider"},
			PreferredIntegration: IntegrationConfigurable,
			IntegrationReference: "https://aider.chat/docs/usage/conventions.html",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Use aider --read AGENTS.md or set read: AGENTS.md in .aider.conf.yml; no automatic loading claim.",
		},
		{
			Name:                 "Amp",
			Description:          "AI-native code editor",
			GlobalConfigPath:     "~/.config/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.config/amp"},
			DetectCommands:       []string{"amp"},
			IntegrationReference: "https://ampcode.com/docs/customize/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Also loads ~/.config/amp/AGENTS.md; avoid duplicating the same instructions in both global paths.",
		},
		{
			Name:                 "Antigravity",
			Description:          "Google cloud IDE with Gemini integration",
			GlobalConfigPath:     "~/.gemini/GEMINI.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.gemini/antigravity-cli"},
			DetectCommands:       []string{"agy"},
			IntegrationReference: "https://antigravity.google/docs/cli/best-practices/",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Current CLI reads root AGENTS.md or GEMINI.md; older IDE installations use rules settings.",
		},
		{
			Name:                 "Autohand",
			Description:          "AI coding assistant",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.autohand", "~/.config/autohand"},
			DetectCommands:       []string{"autohand"},
			IntegrationReference: "https://docs.autohand.ai/working-with-autohand-code/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Project instructions load at conversation start.",
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
			IntegrationReference:   "https://code.claude.com/docs/en/memory",
			ReviewedOn:             "2026-09-05",
			IntegrationNotes:       "Keep a real CLAUDE.md with @AGENTS.md when adding Claude-only instructions.",
		},
		{
			Name:                 "Cline",
			Description:          "Autonomous coding agent for VS Code",
			GlobalConfigPath:     "~/.agents/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/Documents/Cline"},
			DetectCommands:       []string{},
			IntegrationReference: "https://docs.cline.bot/customization/cline-rules",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Root and shared global AGENTS.md are supported; rules can be toggled in Cline.",
		},
		{
			Name:                 "Continue",
			Description:          "Open-source AI code assistant",
			GlobalConfigPath:     "",
			RepoFileName:         ".continue/rules/AGENTS.md",
			ReadsAgentsMD:        false,
			DetectPaths:          []string{"~/.continue"},
			DetectCommands:       []string{"cn", "continue"},
			IntegrationReference: "https://docs.continue.dev/customize/deep-dives/rules",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Create a local Markdown rule in .continue/rules; model/provider config.yaml is not an instruction alias.",
		},
		{
			Name:                 "Crush",
			Description:          "Charm terminal coding agent",
			GlobalConfigPath:     "~/.config/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.config/crush"},
			DetectCommands:       []string{"crush"},
			IntegrationReference: "https://github.com/charmbracelet/crush",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Shared global context is ~/.config/AGENTS.md; Crush-specific context can use ~/.config/crush/CRUSH.md.",
		},
		{
			Name:                 "Cursor",
			Description:          "AI-first code editor",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.cursor"},
			DetectCommands:       []string{"cursor"},
			IntegrationReference: "https://cursor.com/docs/rules",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Root and nested AGENTS.md apply to Agent; global User Rules are configured in the editor.",
		},
		{
			Name:                 "Factory (Droid)",
			Description:          "AI software engineering platform",
			GlobalConfigPath:     "~/.factory/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.factory"},
			DetectCommands:       []string{"droid"},
			IntegrationReference: "https://docs.factory.ai/harness/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Droid reads repository guidance; detection uses the droid command.",
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
			IntegrationReference:   "https://geminicli.com/docs/cli/gemini-md/",
			ReviewedOn:             "2026-09-05",
			IntegrationNotes:       "Set context.fileName to AGENTS.md in Gemini settings to avoid an alias.",
		},
		{
			Name:                 "GitHub Copilot",
			Description:          "AI pair programmer by GitHub",
			GlobalConfigPath:     "~/.copilot/copilot-instructions.md",
			RepoFileName:         ".github/copilot-instructions.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.copilot"},
			DetectCommands:       []string{"copilot"},
			IntegrationReference: "https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Native AGENTS.md in CLI and supported agent clients; .github/copilot-instructions.md remains the cross-feature alias. IDE support varies.",
		},
		{
			Name:                 "Goose",
			Description:          "AI developer agent by Block",
			GlobalConfigPath:     "~/.config/goose/.goosehints",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.config/goose"},
			DetectCommands:       []string{"goose"},
			IntegrationReference: "https://goose-docs.ai/docs/guides/context-engineering/using-goosehints/",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Developer extension required. Defaults load AGENTS.md and .goosehints; avoid duplicate instructions. CONTEXT_FILE_NAMES overrides discovery.",
		},
		{
			Name:                 "Junie",
			Description:          "JetBrains AI coding agent",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.junie"},
			DetectCommands:       []string{},
			IntegrationReference: "https://junie.jetbrains.com/docs/junie-plugin-project-settings.html",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Root AGENTS.md is read unless .junie/AGENTS.md or an explicit Guidelines path overrides it.",
		},
		{
			Name:                 "Kilo Code",
			Description:          "AI coding assistant",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.kilo"},
			DetectCommands:       []string{"kilo"},
			IntegrationReference: "https://kilo.ai/docs/customize/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Root and per-directory AGENTS.md supported; uppercase filename required.",
		},
		{
			Name:                 "Kiro",
			Description:          "AWS agentic IDE",
			GlobalConfigPath:     "~/.kiro/steering/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.kiro"},
			DetectCommands:       []string{"kiro"},
			IntegrationReference: "https://kiro.dev/docs/steering/",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Default agents load AGENTS.md; custom agents must explicitly include steering resources.",
		},
		{
			Name:                 "Codex CLI",
			Description:          "OpenAI command-line coding agent",
			GlobalConfigPath:     "~/.codex/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.codex"},
			DetectCommands:       []string{"codex"},
			IntegrationReference: "https://developers.openai.com/codex/guides/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "CODEX_HOME can override the global directory; AGENTS.override.md takes precedence.",
		},
		{
			Name:                 "OpenClaw",
			Description:          "Personal AI assistant with an agent workspace",
			GlobalConfigPath:     "~/.openclaw/workspace/AGENTS.md",
			RepoFileName:         "",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.openclaw"},
			DetectCommands:       []string{"openclaw"},
			IntegrationReference: "https://docs.openclaw.ai/concepts/agent-workspace",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Workspace instructions, not arbitrary repository discovery. Sandbox seeding rejects aliases outside the workspace.",
		},
		{
			Name:                 "OpenCode",
			Description:          "Terminal-based AI coding assistant",
			GlobalConfigPath:     "~/.config/opencode/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.config/opencode"},
			DetectCommands:       []string{"opencode"},
			IntegrationReference: "https://opencode.ai/v2/docs/instructions",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Global path honors XDG_CONFIG_HOME. V2 reads AGENTS.md; older fallback rules differ.",
		},
		{
			Name:                 "Qwen Code",
			Description:          "Alibaba open-source CLI for agentic coding",
			GlobalConfigPath:     "~/.qwen/QWEN.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.qwen"},
			DetectCommands:       []string{"qwen"},
			IntegrationReference: "https://qwenlm.github.io/qwen-code-docs/en/users/features/memory/",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Current Qwen also reads AGENTS.md; QWEN.md remains supported. Avoid loading duplicate copies.",
		},
		{
			Name:                 "RooCode",
			Description:          "AI coding assistant for VS Code",
			GlobalConfigPath:     "~/.roo/rules/AGENTS.md",
			RepoFileName:         ".roo/rules/rules.md",
			ReadsAgentsMD:        false,
			DetectPaths:          []string{"~/.roo"},
			DetectCommands:       []string{},
			IntegrationReference: "https://roocodeinc.github.io/Roo-Code/features/custom-instructions/",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "An explicit .md file in .roo/rules supplies shared rules; no root AGENTS.md autodiscovery claim.",
		},
		{
			Name:                 "Windsurf",
			Description:          "AI-powered IDE by Codeium",
			GlobalConfigPath:     "",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.codeium"},
			DetectCommands:       []string{"windsurf"},
			IntegrationReference: "https://docs.devin.ai/desktop/cascade/agents-md",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Cascade supports root and nested AGENTS.md; documentation now lives under Devin Desktop.",
		},
		{
			Name:                 "Zed",
			Description:          "High-performance multiplayer code editor",
			GlobalConfigPath:     "~/.config/zed/AGENTS.md",
			RepoFileName:         "AGENTS.md",
			ReadsAgentsMD:        true,
			DetectPaths:          []string{"~/.config/zed"},
			DetectCommands:       []string{"zed"},
			IntegrationReference: "https://zed.dev/docs/ai/instructions",
			ReviewedOn:           "2026-09-05",
			IntegrationNotes:     "Zed Agent uses the first matching project instruction file; external agents use their own loaders.",
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
