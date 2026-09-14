package model

type TaskActionPayload struct {
	Command         string            `json:"command"`
	DeviceOverrides map[string]string `json:"deviceOverrides"`
}
