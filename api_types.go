package main

// API codes

const (
	APIERR_UNKNOWN_USER = 10013
	APIERR_UNAUTHORIZED = 40001
)

// API call error
type APIResponseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type responseError APIResponseError // TODO: get rid of this

