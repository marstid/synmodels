// Package variants provides built-in variant presets for reasoning models.
//
// The Synthetic API does not expose variant definitions. This registry holds
// the variant presets (named sets of request options, e.g. reasoningEffort)
// keyed by the opencode model key (hf:<hugging_face_id>). To add support for a
// new reasoning model, add a single entry to the registry map below.
package variants

// Spec describes the variant-related configuration for a model.
type Spec struct {
	// Reasoning indicates the model supports reasoning/thinking tokens.
	Reasoning bool
	// Interleaved controls how opencode parses interleaved reasoning from
	// streaming responses. Set to true, or a map like {"field":"reasoning_content"}.
	Interleaved any
	// Variants maps variant names to their request option overrides.
	Variants map[string]map[string]any
	// Options provides default request options merged into every request.
	Options map[string]any
}

// registry maps opencode model keys to their variant specs.
var registry = map[string]Spec{
	"hf:zai-org/GLM-5.2": {
		Reasoning: true,
		Variants: map[string]map[string]any{
			"none": {"reasoningEffort": "none"},
			"high": {"reasoningEffort": "high"},
			"max":  {"reasoningEffort": "xhigh"},
		},
	},
}

// Lookup returns the variant spec for the given opencode model key, if one exists.
func Lookup(opencodeKey string) (Spec, bool) {
	spec, ok := registry[opencodeKey]
	return spec, ok
}
