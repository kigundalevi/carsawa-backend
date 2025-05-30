package models

import "time"

// UserBasicRegistrationData represents the minimal registration details for a user.
type UserBasicRegistrationData struct {
	Username    string `json:"username"`    // Required
	Email       string `json:"email"`       // Required
	Password    string `json:"password"`    // Required
	PhoneNumber string `json:"phoneNumber"` // Required
}

// For users, we support three steps: "basic", "otp", and "preferences".
type UserRegistrationRequest struct {
	Step        string                     `json:"step"`                  // "basic", "otp", or "preferences"
	SessionID   string                     `json:"sessionID,omitempty"`   // Provided in "otp" and "preferences" steps
	OTP         string                     `json:"otp,omitempty"`         // Provided in the "otp" step
	BasicData   *UserBasicRegistrationData `json:"basicData,omitempty"`   // Provided in "basic" (and repeated in "otp" if needed)
	Preferences []string                   `json:"preferences,omitempty"` // Provided in "preferences" step
}

// UserRegistrationSession holds all transient data during the multi‑step registration process.
type UserRegistrationSession struct {
	TempID        string                     `json:"tempId" bson:"tempId"`                           // Unique session ID
	BasicData     *UserBasicRegistrationData `json:"basicData,omitempty" bson:"basicData,omitempty"` // Basic registration data
	OTPStatus     string                     `json:"otpStatus" bson:"otpStatus"`                     // "pending" or "verified"
	CreatedAt     time.Time                  `json:"createdAt" bson:"createdAt"`                     // When the session was created
	LastUpdatedAt time.Time                  `json:"lastUpdatedAt" bson:"lastUpdatedAt"`             // When the session was last updated
	Devices       []Device                   `json:"devices,omitempty" bson:"devices,omitempty"`     // Device(s) associated with registration
}
