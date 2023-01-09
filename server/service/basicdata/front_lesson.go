package basicdata

import (
	"github.com/prl26/exam-system/server/global"
	"github.com/prl26/exam-system/server/model/basicdata"
	"github.com/prl26/exam-system/server/model/lessondata"
)

func (lessonService *LessonService) FindLessonDetail(id uint, all bool) (*basicdata.Lesson, error) {
	var lesson basicdata.Lesson
	lesson.ID = id
	if err := global.GVA_DB.Model(&lesson).Preload("Chapters").Find(&lesson).Error; err != nil {
		return nil, err
	}
	if all {
		for _, chapter := range lesson.Chapters {
			if err := global.GVA_DB.Model(&chapter).Preload("Knowledges").Find(&chapter).Error; err != nil {
				return nil, err
			}
		}
	}
	return &lesson, nil
}

func (lessonService *LessonService) FindKnowledge(idUint uint) ([]*lessondata.Knowledge, error) {
	var chapters []*lessondata.Knowledge
	if err := global.GVA_DB.Where("chapter_id=?", idUint).Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}
