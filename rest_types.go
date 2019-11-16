package main

// Generic REST Types

// API call error
type responseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}
