package listing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const nhtsaURL = "https://vpic.nhtsa.dot.gov/api/vehicles/DecodeVINValuesExtended/%s?format=json"

type nhtsaResponse struct {
	Results []nhtsaResult `json:"Results"`
}

type nhtsaResult struct {
	ErrorCode    string `json:"ErrorCode"`
	Make         string `json:"Make"`
	Model        string `json:"Model"`
	ModelYear    string `json:"ModelYear"`
	BodyClass    string `json:"BodyClass"`
	PlantCountry string `json:"PlantCountry"`
}

type NHTSAVerifier struct {
	client *http.Client
}

func NewNHTSAVerifier() *NHTSAVerifier {
	return &NHTSAVerifier{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *NHTSAVerifier) VerifyVIN(ctx context.Context, vin string) (*nhtsaResult, bool, error) {
	url := fmt.Sprintf(nhtsaURL, vin)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("NHTSA API status: %d", resp.StatusCode)
	}

	var apiResp nhtsaResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, false, err
	}
	if len(apiResp.Results) == 0 {
		return nil, false, fmt.Errorf("no results from NHTSA")
	}

	result := apiResp.Results[0]
	return &result, result.ErrorCode == "0", nil
}
