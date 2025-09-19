package model

type Node struct {
	Name      string   `json:"name"`      // 节点名称
	State     string   `json:"state"`     // 节点状态
	Partition []string `json:"partition"` // 分区名称
	Memory    int64    `json:"memory"`    // 内存大小，单位 MB
	CPUs      int64    `json:"cpus"`      // 逻辑CPU
	Socket    int64    `json:"socket"`    // CPU 插槽数
	Cores     int64    `json:"cores"`     // 每个插槽的核心数
	Threads   int64    `json:"threads"`   // 每个核心的线程数
	GPU       string   `json:"gpu"`       // GPU 信息
}

type Nodes []*Node
