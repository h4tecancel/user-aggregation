package response

// Summary represents the total cost from filtered results
// @Description Summary response with total cost calculation
type Summary struct {
	// TotalCost is the sum of prices from filtered records
	TotalCost int64 `json:"total_cost"`
}

// ErrPayload — standard API error shape.
// @Description Returned for all non-2xx responses.
type ErrorPayload struct {
	// Human-readable error message
	Error string `json:"error"`
	// Operation/context where the error occurred (optional)
	Op string `json:"op,omitempty"`
	// HTTP status code (mirrors the response status)
	Status int `json:"status"`
}
