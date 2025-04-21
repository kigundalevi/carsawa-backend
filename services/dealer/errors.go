package dealer

import "fmt"

var (
	ErrDealerNotFound    = fmt.Errorf("dealer not found")
	ErrDealerExists      = fmt.Errorf("dealer already exists")
	ErrInvalidDealerData = fmt.Errorf("invalid dealer data")
	ErrSlugExists        = fmt.Errorf("dealer slug already in use")
)

type DealerValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e DealerValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
