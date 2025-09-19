package model

type Cluster struct {
	Slurmrestd        string `gorm:"slurmrestd"`
	SlurmrestdVersion string `gorm:"slurmrestd_version"`
}
