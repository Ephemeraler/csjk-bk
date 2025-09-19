package model

type Users []User

// User represents a row in user_table.
type User struct {
	CreationTime uint64 `json:"creation_time"`
	ModTime      uint64 `json:"mod_time"`
	Deleted      int8   `json:"deleted"`
	Name         string `json:"name"`
	AdminLevel   int16  `json:"admin_level"`
}
