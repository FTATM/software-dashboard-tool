package model

import (
	"context"
	"math"
	"time"
)

const DeviceScale = 1000

type Device struct {
	DeviceId             int        `json:"deviceId" db:"device_id"`
	DeviceName           string     `json:"deviceName" db:"device_name"`
	Protocol             *string    `json:"protocol" db:"protocol"`
	ValueData            int        `json:"valueData,omitempty" db:"value_data"`
	CreatedAt            time.Time  `json:"-" db:"created_at"`
	UpdatedAt            time.Time  `json:"-" db:"updated_at"`
	UpdatedValueAt       *time.Time `json:"-" db:"updated_value_at"`
	DeletedAt            *time.Time `json:"-" db:"deleted_at"`
	Active               bool       `json:"active" db:"active"`
	LastSeenAt           *time.Time `json:"lastSeenAt,omitempty" db:"last_seen_at"`
	LastAlertTriggeredAt *time.Time `json:"-" db:"last_alert_triggered_at"`

	RefDeviceId *int     `json:"refDeviceId,omitempty" db:"ref_device_id"`
	RawMin      *float64 `json:"rawMin,omitempty" db:"raw_min"`
	RawMax      *float64 `json:"rawMax,omitempty" db:"raw_max"`
	EuMin       *float64 `json:"euMin,omitempty" db:"eu_min"`
	EuMax       *float64 `json:"euMax,omitempty" db:"eu_max"`
}

func (s *Device) IsSame(req Device) bool {
	if s == nil {
		return false
	}

	return s.DeviceId == req.DeviceId &&
		s.DeviceName == req.DeviceName &&
		s.Active == req.Active &&
		ptrEqual(s.Protocol, req.Protocol) &&
		ptrEqual(s.RefDeviceId, req.RefDeviceId) &&
		ptrEqual(s.RawMin, req.RawMin) &&
		ptrEqual(s.RawMax, req.RawMax) &&
		ptrEqual(s.EuMin, req.EuMin) &&
		ptrEqual(s.EuMax, req.EuMax)
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func Remap(value float64, rawMin, rawMax, euMin, euMax *float64) float64 {
	if rawMin == nil || rawMax == nil || euMin == nil || euMax == nil || *rawMax == *rawMin {
		return value
	}
	// Clamp raw value to configured boundaries
	clamped := math.Max(*rawMin, math.Min(*rawMax, value))
	ratio := (clamped - *rawMin) / (*rawMax - *rawMin)
	return *euMin + ratio*(*euMax-*euMin)
}

type DeviceDetail struct {
	DeviceId    int        `json:"deviceId"`
	DeviceName  string     `json:"deviceName"`
	Protocol    *string    `json:"protocol,omitempty" db:"protocol"`
	ValueData   int        `json:"valueData,omitempty"`
	Active      bool       `json:"active,omitempty"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty"`
	RefDeviceId *int       `json:"refDeviceId,omitempty"`
	RawMin      *float64   `json:"rawMin,omitempty"`
	RawMax      *float64   `json:"rawMax,omitempty"`
	EuMin       *float64   `json:"euMin,omitempty"`
	EuMax       *float64   `json:"euMax,omitempty"`
}

type DeviceCreate struct {
	DeviceName  string   `json:"deviceName"`
	Protocol    *string  `json:"protocol"`
	Active      bool     `json:"active"`
	RefDeviceId *int     `json:"refDeviceId"`
	RawMin      *float64 `json:"rawMin"`
	RawMax      *float64 `json:"rawMax"`
	EuMin       *float64 `json:"euMin"`
	EuMax       *float64 `json:"euMax"`
}

type DeviceUpdate struct {
	DeviceId    int      `json:"deviceId"`
	DeviceName  string   `json:"deviceName"`
	Protocol    *string  `json:"protocol"`
	Active      bool     `json:"active"`
	OldName     string   `json:"-"`
	RefDeviceId *int     `json:"refDeviceId"`
	RawMin      *float64 `json:"rawMin"`
	RawMax      *float64 `json:"rawMax"`
	EuMin       *float64 `json:"euMin"`
	EuMax       *float64 `json:"euMax"`
}

type ChartDeviceData struct {
	DeviceName     string     `json:"deviceName"`
	ValueData      float64    `json:"valueData"`
	UpdatedValueAt *time.Time `json:"updatedValueAt"`
	RawMin         *float64   `json:"-" db:"raw_min"`
	RawMax         *float64   `json:"-" db:"raw_max"`
	EuMin          *float64   `json:"-" db:"eu_min"`
	EuMax          *float64   `json:"-" db:"eu_max"`
}

type ChartData struct {
	DeviceData map[int]ChartDeviceData `json:"deviceData"`
}

type ChartHistoryData struct {
	DeviceName string       `json:"name"`
	Data       [][2]float64 `json:"data"`
}

type TriggerCommandReq struct {
	DeviceIds  []int       `json:"deviceIds"`
	IsGroup    bool        `json:"isGroup"`
	TaskAction DynamicJSON `json:"taskAction"`
}

type CommandDeviceInfo struct {
	DeviceId      int     `db:"device_id"`
	DeviceName    string  `db:"device_name"`
	Protocol      *string `db:"protocol"`
	GroupId       *int    `db:"group_id"`
	GroupProtocol *string `db:"group_protocol"`
}

type DeviceRepository interface {
	GetById(ctx context.Context, id int) (*Device, error)
	GetAll(ctx context.Context, active bool) ([]Device, error)
	GetAllName(ctx context.Context, active bool) ([]Device, error)
	Create(ctx context.Context, device []Device) error
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, deviceId int) error
	GetProtocolType(ctx context.Context) ([]string, error)
	GetByIdChartDeviceData(ctx context.Context, id int) (ChartDeviceData, error)
	GetAggregatedData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, bucketInterval string) ([]DeviceDataLog, error)
	GetRawData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, limit int) ([]DeviceDataLog, error)
	CountData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time) (int, error)
	GetDeviceGroupById(ctx context.Context, deviceGroupId int) (*DeviceGroup, error)
	GetDeviceIdByDeviceGroupIds(ctx context.Context, deviceGroupId []int) (map[int][]int, error)
	GetAllDeviceGroup(ctx context.Context) ([]DeviceGroup, error)
	CreateGroup(ctx context.Context, deviceGroup *DeviceGroup) error
	UpdateGroup(ctx context.Context, deviceGroup *DeviceGroup) error
	DeleteGroup(ctx context.Context, deviceGroupId int) error
	CreateGroupMap(ctx context.Context, groupId int, deviceIds []int) error
	DeleteGroupMap(ctx context.Context, groupId int, deviceIds []int) error
	GetByIds(ctx context.Context, id []int, active bool) ([]Device, error)
	GetDeviceForCommandByIds(ctx context.Context, deviceIds []int) ([]CommandDeviceInfo, error)
	GetDeviceProtocol(ctx context.Context, deviceId int) (*string, error)
}

type DeviceService interface {
	GetAllDeviceDetail(ctx context.Context) ([]DeviceDetail, error)
	CreateDevice(ctx context.Context, createDeviceReq []DeviceCreate, authUserId int) error
	UpdateDevice(ctx context.Context, updateDeviceReq *DeviceUpdate, authUserId int) error
	DeleteDevice(ctx context.Context, deviceId, authUserId int) (*Device, error)
	GetProtocolType(ctx context.Context) ([]string, error)
	GetAllDeviceName(ctx context.Context) ([]DeviceDetail, error)
	StartPublic(ctx context.Context)
	AddClient(deviceID string, clientChan chan ChartData)
	RemoveClient(deviceID string, clientChan chan ChartData)
	GetChartHistory(ctx context.Context, deviceId []int, maxPoints int, from, to time.Time) (map[int]ChartHistoryData, error)
	GetDeviceGroupDetail(ctx context.Context) ([]DeviceGroupDetail, error)
	CreateDeviceGroup(ctx context.Context, createDeviceG CreateDeviceGroup, authUserId int) error
	UpdateDeviceGroup(ctx context.Context, updateDeviceG *UpdateDeviceGroup, authUserId int) error
	DeleteDeviceGroup(ctx context.Context, deviceGroupId, authUserId int) (*DeviceGroup, error)

	GetDeviceForCommandByIds(ctx context.Context, deviceIds []int) ([]CommandDeviceInfo, error)
}
