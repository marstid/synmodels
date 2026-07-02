// Package types defines the data structures used throughout the application.
package types

// Model represents a single model from the API response.
type Model struct {
	ID                          string   `json:"id"`
	HuggingFaceID               string   `json:"hugging_face_id"`
	Name                        string   `json:"name"`
	Object                      string   `json:"object"`
	Created                     int64    `json:"created"`
	OwnedBy                     string   `json:"owned_by"`
	Provider                    string   `json:"provider"`
	AlwaysOn                    bool     `json:"always_on"`
	InputModalities             []string `json:"input_modalities"`
	OutputModalities            []string `json:"output_modalities"`
	ContextLength               int      `json:"context_length"`
	MaxOutputLength             int      `json:"max_output_length"`
	SupportedFeatures           []string `json:"supported_features"`
	SupportedSamplingParameters []string `json:"supported_sampling_parameters"`
	Quantization                string   `json:"quantization"`
	Pricing                     Pricing  `json:"pricing"`
}

// Pricing represents the pricing information for a model.
type Pricing struct {
	Prompt           string `json:"prompt"`
	Completion       string `json:"completion"`
	Image            string `json:"image"`
	Request          string `json:"request"`
	InputCacheReads  string `json:"input_cache_reads"`
	InputCacheWrites string `json:"input_cache_writes"`
}

// ModelsResponse represents the API response containing a list of models.
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// SelectedModel represents a model with its selection state.
type SelectedModel struct {
	Model    Model
	Selected bool
}
