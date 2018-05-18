package main

type Plan struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// nolint: megacheck, deadcode
type InstanceForm struct {
	Name string `json:"name"`
	Plan string `json:"plan"`
	Team string `json:"team"`
	User string `json:"user"`
	Tag  []string
}

// nolint: megacheck, deadcode
type BindUnitForm struct {
	AppHost  string `json:"apphost"`
	AppName  string `json:"appname"`
	UnitHost string `json:"unithost"`
}

type BindAppForm struct {
	AppHost string `json:"apphost"`
	AppName string `json:"appname"`
}
