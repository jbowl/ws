package types

type BreweryResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BreweryType string `json:"brewery_type"`

	Street string `json:"street"`
	//	address_2: null,
	//	address_3: null,
	City  string `json:"city"`
	State string `json:"state"`

	CountryProvince string `json:"CountryProvince"`
	PostalCode      string `json:"postal_code"`
	Country         string `json:"country"`
	Longitude       string `json:"longitude"`
	Latitude        string `json:"latitude"`
	Phone           string `json:"phone"`

	Website string `json:"website_url"`
	Updated string `json:"updated_at"`
	Created string `json:"created_at"`
	//	updated_at: "2018-08-23T23:24:11.758Z",
	//	created_at: "2018-08-23T23:24:11.758Z"
	//	MapURL string `json:"mapurl"`
}

// ProblemDetails - see RFC 7807 Problem Details
//https://tools.ietf.org/html/rfc7807
//Error responses will have each of the following keys:
//detail (string) - A human-readable description of the specific error.
//type (string) - a URL to a document describing the error condition (optional, and "about:blank" is assumed if none is provided; should resolve to a human-readable document).
//title (string) - A short, human-readable title for the general error type; the title should not change for given types.
//status (number) - Conveying the HTTP status code; this is so that all information is in one place, but also to correct for changes in the status code due to the usage of proxy servers. The status member, if present, is only advisory as generators MUST use the same status code in the actual HTTP response to assure that generic HTTP software that does not understand this format still behaves correctly.
//instance (string) - This optional key may be present, with a unique URI for the specific error; this will often point to an error log for that specific response.
type ProblemDetails struct {
	Detail   string
	Type     string
	Title    string
	Status   int
	Instance string
}
