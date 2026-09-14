package model

import "time"

// DeviceGroup represents the main group entity
type DeviceGroup struct {
	GroupId     int       `json:"groupId" db:"group_id"`
	GroupName   string    `json:"groupName" db:"group_name"`
	Protocol    *string   `json:"protocol" db:"protocol"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

func (s *DeviceGroup) IsSame(req DeviceGroup) bool {
	if s == nil {
		return false
	}

	return s.GroupId == req.GroupId &&
		s.GroupName == req.GroupName &&
		s.Description == req.Description &&
		s.Protocol == req.Protocol
}

// DeviceGroupMapping represents the payload to add/remove a device from a group
type DeviceGroupMap struct {
	GroupId  int `json:"groupId" db:"group_id"`
	DeviceId int `json:"deviceId" db:"device_id"`
}

// DeviceGroupResponse is useful when returning a group along with its devices to the frontend
type DeviceGroupDetail struct {
	DeviceGroup
	DeviceIds []int `json:"deviceIds"`
}

type DeviceGroupData struct {
	GroupId   int    `json:"groupId" db:"group_id"`
	GroupName string `json:"groupName" db:"group_name"`
	Protocol  string `json:"protocol" db:"protocol"`
	DeviceIds []int  `json:"deviceIds" db:"device_id_s"`
}

type CreateDeviceGroup struct {
	GroupName   string  `json:"groupName"`
	Description string  `json:"description"`
	Protocol    *string `json:"protocol"`
	DeviceIds   []int   `json:"deviceIds"`
}

type UpdateDeviceGroup struct {
	GroupId     int     `json:"groupId"`
	GroupName   string  `json:"groupName"`
	Protocol    *string `json:"protocol"`
	Description string  `json:"description"`
	DeviceIds   []int   `json:"deviceIds"`
	OldName     string  `json:"-"`
}
