package models

import (
	"time"

	"github.com/google/uuid"
)

// UserInfo represents user subscription information
// @Description User subscription information with service details and pricing
type UserInfo struct {
	// ServiceName is the name of the subscribed service
	ServiceName string `json:"service_name"`
	// Price is the subscription price (always integer)
	Price int64 `json:"price"`
	// UserID is the unique identifier of the user
	UserID uuid.UUID `json:"user_id"`
	// StartDate is when the subscription begins
	StartDate time.Time `json:"start_date"`
	// EndDate is when the subscription ends
	EndDate time.Time `json:"end_date"`
}


// UpdateUserInfo is universal model. It responds to both requests and responses
// @Description User subscription update fields (all fields are optional, except UserID,It does not need to be filled out to submit a request.)
type UpdateUserInfo struct {
	// !!! UserID is the unique identifier of the user(not for request! ONLY RESPONSE)
	UserID uuid.UUID `json:"user_id"`
	// Price is the updated subscription price (optional, always integer)
	Price *int64 `json:"price,omitempty"`
	// EndDate is the updated subscription end date (optional)
	EndDate *time.Time `json:"end_date,omitempty"`
}
