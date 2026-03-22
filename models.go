package main

type JobPayload struct {
	Task    string `json:"task"`
	Value   string `json:"value"`
	Seconds int    `json:"seconds"`
}
