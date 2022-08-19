// 自动生成模板Professional
package basicdata

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	
)

// Professional 结构体
type Professional struct {
      global.GVA_MODEL
      Name  string `json:"name" form:"name" gorm:"column:name;comment:专业名称;size:255;"`
      College_id  *int `json:"college_id" form:"college_id" gorm:"column:college_id;comment:学院id;"`
}


// TableName Professional 表名
func (Professional) TableName() string {
  return "bas_professional"
}

