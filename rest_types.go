package main

type responseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type responseGetGateway struct {
        URL string `json:"url"`
}

