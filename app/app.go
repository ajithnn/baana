package app

type App struct {
	Name           string   `json:"app_name"`
	Path           string   `json:"path"`
	ImportPath     string   `json:"import_path"`
	ControllerList []string `json:"controller_list"`
}
