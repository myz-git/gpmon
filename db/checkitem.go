// gpmon/db/checkitem.go
package db

// CheckItem 监控检查项（与各数据库驱动无关的公共类型）
type CheckItem struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string
	Frequency int
	IsEnable  int
}
