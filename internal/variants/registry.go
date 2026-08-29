// Package variants provides built-in overrides for reasoning models.
//
// Variant presets (named sets of request options, e.g. reasoningEffort) are now
// generated dynamically from the API's reasoning_parameters.efforts field (see
// internal/config/generator.go). This registry only holds per-model overrides
// that the API does not expose, keyed by the opencode model key (hf:<hugging_face_id>).
//
// Currently the only such override is `interleaved`, which controls how opencode
// parses interleaved reasoning from streaming responses (e.g. Kimi-K3 emits it
// in a `reasoning_content` field). To add an override for a new model, add a
// single entry to the registry map below.
package variants

// Spec describes the variant-related configuration overrides for a model.
type Spec struct {
	// Reasoning overrides the reasoning flag for the model. When set, forces the
	// model's reasoning flag regardless of supported_features. Leave as false to
	// let the generator detect it from supported_features.
	Reasoning bool
	// Interleaved controls how opencode parses interleaved reasoning from
	// streaming responses. Set to true, or a map like {"field":"reasoning_content"}.
	Interleaved any
	// Variants overrides the dynamically generated variant presets for the
	// model. When nil, variants are generated from reasoning_parameters.efforts.
	Variants map[string]map[string]any
	// Options provides default request options merged into every request.
	Options map[string]any
}

// registry maps opencode model keys to their override specs.
var registry = map[string]Spec{
	"hf:moonshotai/Kimi-K3": {
		Interleaved: map[string]any{"field": "reasoning_content"},
	},
}

// Lookup returns the override spec for the given opencode model key, if one exists.
func Lookup(opencodeKey string) (Spec, bool) {
	spec, ok := registry[opencodeKey]
	return spec, ok
}
