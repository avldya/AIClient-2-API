package antigravity

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

// GenerateRequestID generates a random request ID in the format "agent-{uuid}"
func GenerateRequestID() string {
	return "agent-" + uuid.New().String()
}

// GenerateSessionID generates a random session ID with format "-{random_number}"
// The random number is between 9 and 18 digits
func GenerateSessionID() string {
	// Generate a random number between 0 and 9000000000000000000
	n := rand.Int63n(9000000000000000000)
	return fmt.Sprintf("-%d", n)
}

// GenerateProjectID generates a random project ID in the format "{adj}-{noun}-{random5}"
func GenerateProjectID() string {
	adjectives := []string{"useful", "bright", "swift", "calm", "bold"}
	nouns := []string{"fuze", "wave", "spark", "flow", "core"}

	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	randomPart := strings.ToLower(uuid.New().String()[:5])

	return fmt.Sprintf("%s-%s-%s", adj, noun, randomPart)
}

// Alias2ModelName converts an alias to the real model name
func Alias2ModelName(modelName string) string {
	if realName, ok := ModelAliasMap[modelName]; ok {
		return realName
	}
	return modelName
}

// ModelName2Alias converts a real model name to its alias
func ModelName2Alias(modelName string) string {
	if alias, ok := ModelNameMap[modelName]; ok {
		return alias
	}
	return modelName
}
