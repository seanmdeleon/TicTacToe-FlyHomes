package apiresources

// Response is the Response sent to from all endpoints
// The API Client should check the Error to see if the request was valid
// If no Error exists, access the Data
type Response struct {
	// ErrorMessage holds all error messages encountered in an http call. If no error exists, it will be nil
	ErrorMessage *string `json:"errorMessage"`

	// Data holds the response from the CRUD operation on the database
	Data map[string]interface{} `json:"data"`
}
