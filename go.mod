module github.com/snapsynapse/agentlink

go 1.23

require (
	github.com/spf13/cobra v1.10.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

// Retract every release before the final security fix in v0.6.1: in those
// versions a cloned repository's .agentlink.yaml could create symlinks
// outside the repository.
retract [v0.0.0, v0.6.0]
