package registry

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

// Documentation renders the complete support listing from the runtime registry.
func Documentation() string {
	var b strings.Builder
	b.WriteString("<div class=\"support-table\" role=\"region\" aria-label=\"AI tool support\" tabindex=\"0\">\n<table>\n<caption>Instruction integrations and vendor references</caption>\n<thead><tr><th scope=\"col\">Tool / mode</th><th scope=\"col\">Instruction paths</th><th scope=\"col\">Requirements and evidence</th></tr></thead>\n<tbody>\n")
	tools := All()
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	path := func(s string) string {
		if s == "" {
			return "Not listed"
		}
		return "<code>" + html.EscapeString(s) + "</code>"
	}
	for _, tool := range tools {
		fmt.Fprintf(&b, "<tr><th scope=\"row\">%s<br><small>%s</small></th><td>Project: %s<br>Global: %s</td><td>%s <a href=\"%s\">Vendor reference for %s</a> (reviewed %s).</td></tr>\n",
			html.EscapeString(tool.Name), html.EscapeString(string(tool.AgentsMDIntegration())), path(tool.RepoFileName), path(tool.GlobalConfigPath), html.EscapeString(tool.IntegrationNotes), html.EscapeString(tool.IntegrationReference), html.EscapeString(tool.Name), html.EscapeString(tool.ReviewedOn))
	}
	b.WriteString("</tbody>\n</table>\n</div>\n")
	return b.String()
}
