package skills

import "strings"

// Start SMALL. You will grow this list later.
var KnownSkills = []string{
	"go", "golang",
	"python",
	"java",
	"javascript",
	"typescript",
	"react",
	"node",
	"sql",
	"postgres",
	"docker",
	"aws",
	"git",
	"linux",
}

func Extract(text string) []string {
	text = strings.ToLower(text)

	found := make(map[string]bool)

	for _, skill := range KnownSkills {
		if strings.Contains(text, skill) {
			found[normalize(skill)] = true
		}
	}

	out := []string{}
	for s := range found {
		out = append(out, s)
	}

	return out
}

func normalize(skill string) string {
	switch skill {
	case "golang":
		return "go"
	case "postgres":
		return "postgresql"
	default:
		return skill
	}
}
