// Package types defines the data structures used throughout the application.
package types

// Model represents a single model from the API response.
type Model struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	OwnedBy           string   `json:"owned_by"`
	InputModalities   []string `json:"input_modalities"`
	OutputModalities  []string `json:"output_modalities"`
	ContextLength     int      `json:"context_length"`
	MaxOutputLength   int      `json:"max_output_length"`
	SupportedFeatures []string `json:"supported_features"`
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
