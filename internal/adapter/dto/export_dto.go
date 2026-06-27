package dto

// ExportCSVRequest CSV 导出请求
type ExportCSVRequest struct {
	// Filename 文件名；为空时后端使用默认文件名
	Filename string `json:"filename"`
	// Content CSV 文件内容
	Content string `json:"content"`
}

// ExportCSVResult CSV 导出结果
type ExportCSVResult struct {
	// Saved 是否已保存；用户取消选择目录时为 false
	Saved bool `json:"saved"`
	// Path 保存后的完整文件路径
	Path string `json:"path"`
}
