package skills

import (
	"regexp"
)

// SkillDef holds the regex pattern and the canonical name
type SkillDef struct {
	Name    string
	Pattern *regexp.Regexp
}

var skillDefs []SkillDef

func init() {
	// Helper to create word-boundary regex
	// e.g. "go" -> `(?i)\bgo\b`
	// Special characters in names (like C++) need escaping if they aren't handled
	makePattern := func(name string, pattern string) SkillDef {
		if pattern == "" {
			pattern = regexp.QuoteMeta(name)
		}
		// Case insensitive, word boundaries
		regex := regexp.MustCompile(`(?i)\b` + pattern + `\b`)
		return SkillDef{Name: name, Pattern: regex}
	}

	rawSkills := []struct {
		Name    string
		Pattern string // Optional custom regex pattern (without \b wrappers)
	}{
		// Languages
		{"Go", "go|golang"},
		{"Python", ""},
		{"Java", ""},
		{"JavaScript", "js|javascript"},
		{"TypeScript", "ts|typescript"},
		{"C++", `c\+\+`},
		{"C#", `c#|csharp`},
		{"C", `c`}, // Dangerous, but \bC\b handles most cases
		{"Ruby", ""},
		{"PHP", ""},
		{"Swift", ""},
		{"Kotlin", ""},
		{"Rust", ""},
		{"Dart", ""},
		{"Scala", ""},
		{"Perl", ""},
		{"Haskell", ""},
		{"Lua", ""},
		{"R", `r`},

		// Frontend
		{"React", "react|reactjs"},
		{"Angular", "angular|angularjs"},
		{"Vue.js", "vue|vuejs|vue\\.js"},
		{"Svelte", ""},
		{"Next.js", "nextjs|next\\.js"},
		{"HTML", ""},
		{"CSS", ""},
		{"Tailwind CSS", "tailwind|tailwind css"},
		{"Bootstrap", ""},
		{"Redux", ""},
		{"Webpack", ""},

		// Backend / Frameworks
		{"Node.js", "node|nodejs|node\\.js"},
		{"Express", "express|expressjs"},
		{"Django", ""},
		{"Flask", ""},
		{"Spring Boot", "spring|spring boot"},
		{"ASP.NET", "asp\\.net|dotnet|\\.net"},
		{"Ruby on Rails", "rails|ruby on rails"},
		{"FastAPI", ""},
		{"GraphQL", ""},

		// Database
		{"SQL", ""},
		{"PostgreSQL", "postgres|postgresql"},
		{"MySQL", ""},
		{"MongoDB", "mongo|mongodb"},
		{"Redis", ""},
		{"Elasticsearch", ""},
		{"SQLite", ""},
		{"Oracle", ""},
		{"DynamoDB", ""},
		{"Cassandra", ""},

		// DevOps / Cloud
		{"AWS", "aws|amazon web services"},
		{"Azure", ""},
		{"Google Cloud", "gcp|google cloud"},
		{"Docker", ""},
		{"Kubernetes", "k8s|kubernetes"},
		{"Jenkins", ""},
		{"Git", ""},
		{"GitHub", ""},
		{"GitLab", ""},
		{"Terraform", ""},
		{"Ansible", ""},
		{"CircleCI", ""},
		{"Linux", ""},
		{"Bash", ""},

		// Mobile
		{"React Native", "react native"},
		{"Flutter", ""},
		{"Android", ""},
		{"iOS", ""},

		// AI / ML
		{"Machine Learning", "machine learning|ml"},
		{"Deep Learning", ""},
		{"TensorFlow", ""},
		{"PyTorch", ""},
		{"Scikit-learn", ""},
		{"Pandas", ""},
		{"NumPy", ""},
		{"OpenCV", ""},
		{"Keras", ""},
		{"NLP", "nlp|natural language processing"},
	}

	for _, s := range rawSkills {
		skillDefs = append(skillDefs, makePattern(s.Name, s.Pattern))
	}
}

// Extract returns a list of unique canonical skill names found in the text.
func Extract(text string) []string {
	// Normalize checks slightly, though regex handles case insensitivity
	// We might want to replace newlines with spaces to help regex running across lines if needed,
	// but \b usually works fine.

	foundMap := make(map[string]bool)

	for _, def := range skillDefs {
		if def.Pattern.MatchString(text) {
			foundMap[def.Name] = true
		}
	}

	var results []string
	for name := range foundMap {
		results = append(results, name)
	}
	return results
}
