package basicdata

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/basicdata"
	basicdataReq "github.com/flipped-aurora/gin-vue-admin/server/model/basicdata/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type TermService struct {
}

// CreateTerm 创建Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService) CreateTerm(term basicdata.Term) (err error) {
	err = global.GVA_DB.Create(&term).Error
	return err
}

// DeleteTerm 删除Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService)DeleteTerm(term basicdata.Term) (err error) {
	err = global.GVA_DB.Delete(&term).Error
	return err
}

// DeleteTermByIds 批量删除Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService)DeleteTermByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]basicdata.Term{},"id in ?",ids.Ids).Error
	return err
}

// UpdateTerm 更新Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService)UpdateTerm(term basicdata.Term) (err error) {
	err = global.GVA_DB.Save(&term).Error
	return err
}

// GetTerm 根据id获取Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService)GetTerm(id uint) (term basicdata.Term, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&term).Error
	return
}

// GetTermInfoList 分页获取Term记录
// Author [piexlmax](https://github.com/piexlmax)
func (termService *TermService)GetTermInfoList(info basicdataReq.TermSearch) (list []basicdata.Term, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&basicdata.Term{})
    var terms []basicdata.Term
    // 如果有条件搜索 下方会自动创建搜索语句
    if info.Name != "" {
        db = db.Where("name = ?",info.Name)
    }
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }
	err = db.Limit(limit).Offset(offset).Find(&terms).Error
	return  terms, total, err
}
