package model

type Menu struct {
	MenuId   int    `json:"menuId" db:"menu_id"`
	MenuName string `json:"menuName" db:"menu_name"`
	ParentId int    `json:"parentId" db:"parent_id"`
}

type SubMenu struct {
	MenuID           int      `json:"menuId"`
	MenuName         string   `json:"menuName"`
	AvailableActions []Action `json:"availableActions"`
}

type MainMenu struct {
	MenuID           int       `json:"menuId" db:"menu_id"`
	MenuName         string    `json:"menuName" db:"menu_name"`
	AvailableActions []Action  `json:"availableActions" db:"available_actions"`
	SubMenus         []SubMenu `json:"submenus" db:"submenus"`
}
