package model

type Action struct {
	ActionId   int    `json:"actionId" db:"action_id"`
	ActionName string `json:"actionName" db:"action_name"`
}
