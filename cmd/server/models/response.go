package models

type Response struct {
	Status string `json:"status"`
}

type ResponseError struct {
	Response
	Error string `json:"error"`
}

type ResponseSuccess struct {
	Response
	Data interface{} `json:"data"`
}
