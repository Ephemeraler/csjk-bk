package model

type QoS struct {
	CreationTime          uint64  `gorm:"column:creation_time" json:"creation_time"`
	ModTime               uint64  `gorm:"column:mod_time" json:"mod_time"`
	Deleted               int8    `gorm:"column:deleted" json:"deleted"`
	ID                    int32   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name                  string  `gorm:"column:name;unique" json:"name"`
	Description           string  `gorm:"column:description" json:"description"`
	Flags                 uint32  `gorm:"column:flags" json:"flags"`
	GraceTime             uint32  `gorm:"column:grace_time" json:"grace_time"`
	MaxJobsPA             int32   `gorm:"column:max_jobs_pa" json:"max_jobs_pa"`
	MaxJobsPerUser        int32   `gorm:"column:max_jobs_per_user" json:"max_jobs_per_user"`
	MaxJobsAccruePA       int32   `gorm:"column:max_jobs_accrue_pa" json:"max_jobs_accrue_pa"`
	MaxJobsAccruePU       int32   `gorm:"column:max_jobs_accrue_pu" json:"max_jobs_accrue_pu"`
	MinPrioThresh         int32   `gorm:"column:min_prio_thresh" json:"min_prio_thresh"`
	MaxSubmitJobsPA       int32   `gorm:"column:max_submit_jobs_pa" json:"max_submit_jobs_pa"`
	MaxSubmitJobsPerUser  int32   `gorm:"column:max_submit_jobs_per_user" json:"max_submit_jobs_per_user"`
	MaxTresPA             string  `gorm:"column:max_tres_pa" json:"max_tres_pa"`
	MaxTresPJ             string  `gorm:"column:max_tres_pj" json:"max_tres_pj"`
	MaxTresPN             string  `gorm:"column:max_tres_pn" json:"max_tres_pn"`
	MaxTresPU             string  `gorm:"column:max_tres_pu" json:"max_tres_pu"`
	MaxTresMinsPJ         string  `gorm:"column:max_tres_mins_pj" json:"max_tres_mins_pj"`
	MaxTresRunMinsPA      string  `gorm:"column:max_tres_run_mins_pa" json:"max_tres_run_mins_pa"`
	MaxTresRunMinsPU      string  `gorm:"column:max_tres_run_mins_pu" json:"max_tres_run_mins_pu"`
	MinTresPJ             string  `gorm:"column:min_tres_pj" json:"min_tres_pj"`
	MaxWallDurationPerJob int32   `gorm:"column:max_wall_duration_per_job" json:"max_wall_duration_per_job"`
	GrpJobs               int32   `gorm:"column:grp_jobs" json:"grp_jobs"`
	GrpJobsAccrue         int32   `gorm:"column:grp_jobs_accrue" json:"grp_jobs_accrue"`
	GrpSubmitJobs         int32   `gorm:"column:grp_submit_jobs" json:"grp_submit_jobs"`
	GrpTres               string  `gorm:"column:grp_tres" json:"grp_tres"`
	GrpTresMins           string  `gorm:"column:grp_tres_mins" json:"grp_tres_mins"`
	GrpTresRunMins        string  `gorm:"column:grp_tres_run_mins" json:"grp_tres_run_mins"`
	GrpWall               int32   `gorm:"column:grp_wall" json:"grp_wall"`
	Preempt               string  `gorm:"column:preempt" json:"preempt"`
	PreemptMode           int32   `gorm:"column:preempt_mode" json:"preempt_mode"`
	PreemptExemptTime     uint32  `gorm:"column:preempt_exempt_time" json:"preempt_exempt_time"`
	Priority              uint32  `gorm:"column:priority" json:"priority"`
	UsageFactor           float64 `gorm:"column:usage_factor" json:"usage_factor"`
	UsageThres            float64 `gorm:"column:usage_thres" json:"usage_thres"`
	LimitFactor           float64 `gorm:"column:limit_factor" json:"limit_factor"`
}

// Qoses is a slice of Qos.
type QoSes []QoS
