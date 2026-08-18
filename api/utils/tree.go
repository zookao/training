package utils

// TreeNode 通用树接口
type TreeNode interface {
	GetID() uint
	GetParentID() uint
	SetChildren(children interface{})
}

// BuildMenuTree 构建菜单树（专用于 admin.Menu，避免反射）
// 见 service 层使用：传入扁平切片返回树。
// 这里直接采用具体类型实现的辅助，见 service/admin/menu.go
