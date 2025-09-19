package model

type Accounts []Account

type Account struct {
	CreationTime uint64 `gorm:"column:creation_time" json:"creation_time"`
	ModTime      uint64 `gorm:"column:mod_time" json:"mod_time"`
	Deleted      int8   `gorm:"column:deleted" json:"deleted"`
	Flags        uint32 `gorm:"column:flags" json:"flags"`
	Name         string `gorm:"column:name;primaryKey" json:"name"`
	Description  string `gorm:"column:description" json:"description"`
	Organization string `gorm:"column:organization" json:"organization"`
}
