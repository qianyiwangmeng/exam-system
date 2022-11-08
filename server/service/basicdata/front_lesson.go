package basicdata

import (
	"github.com/prl26/exam-system/server/global"
	"github.com/prl26/exam-system/server/model/basicdata"
)

func (lessonService *LessonService) FindLessonDetail(id uint) (*basicdata.Lesson, error) {
	var lesson basicdata.Lesson
	lesson.ID = id
	if err := global.GVA_DB.Model(&lesson).Preload("Chapters").Find(&lesson).Error; err != nil {
		return nil, err
	}
	for _, chapter := range lesson.Chapters {
		if err := global.GVA_DB.Model(&chapter).Preload("Knowledges").Find(&chapter).Error; err != nil {
			return nil, err
		}
	}
	return &lesson, nil
}
