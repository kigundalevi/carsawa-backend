package models

type DealerRegistration struct {
	CompanyName string   `json:"companyName"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Location    Location `json:"location"`
}

type TradeInFilter struct {
	UserID     string
	DealerID   string
	MinYear    int
	MaxMileage int
	Status     string
}

type VerificationDocument struct {
	DocumentType string `json:"documentType"` // "license", "tax-certificate"
	FileURL      string `json:"fileUrl"`
}
