package questionBank

import (
	"context"
	"fmt"
	"github.com/prl26/exam-system/server/global"
	"github.com/prl26/exam-system/server/model/questionBank/enum/questionType"
	"github.com/prl26/exam-system/server/model/teachplan"
	teachplanResp "github.com/prl26/exam-system/server/model/teachplan/response"
	"time"
)

type PracticeService struct {
}

func (p PracticeService) FindTheLatestRecord(lessonId, studentId uint) *teachplan.PracticeRecord {
	t := teachplan.PracticeRecord{}
	global.GVA_DB.Where("lesson_id=? and student_id=?", lessonId, studentId).Order("begin_time desc").First(&t)
	return &t
}

func (p PracticeService) CreatePracticeRecord(r *teachplan.PracticeRecord) error {
	return global.GVA_DB.Create(r).Error
}

func (p PracticeService) FindTheLatestItemId(lessonId, studentId uint) *uint {
	result := new(uint)
	global.GVA_DB.Model(&teachplan.PracticeItem{}).Select("id").Where("lessonId=? and studentId=?", lessonId, studentId).Order("answer_time desc").Find(result)
	return result
}

func (p PracticeService) RemoveRecord(id uint) {
	global.GVA_DB.Where("id=?", id).Delete(&teachplan.PracticeRecord{})
}

func (p PracticeService) FindItemCount(lessonId, studentId uint, beginIndex, endIndex *uint) (result int64) {
	global.GVA_DB.Where("id>=? && id<=? && lessonId=? and studentId=?", beginIndex, endIndex, lessonId, studentId).Count(&result)
	return
}

func (p PracticeService) UpdatePracticeRecord(lessonId, studentId uint) {
	record := p.FindTheLatestRecord(lessonId, studentId)
	if record.ID != 0 {
		a := 0
		global.GVA_DB.Raw("update tea_practice_record a \njoin (select count(*) as count\nfrom tea_practice_item\nwhere tea_practice_item.record_id=?) b\nset a.question_count=b.count,a.deleted_at=if(b.count=0,NOW(),null) \nwhere a.id=?", record.ID, record.ID).Scan(&a)
	}
	return
}

func (p PracticeService) CreatePracticeItem(questionType questionType.QuestionType, questionId, lessonId, studentId uint, score uint) {
	record := p.FindTheLatestRecord(lessonId, studentId)
	if record.ID != 0 {
		a := 0
		global.GVA_DB.Raw("INSERT INTO tea_practice_item(student_id,question_type,lesson_id,question_id,record_id,score,commit_time) VALUES (?,?,?,?,?,?,now())\n  ON DUPLICATE KEY UPDATE commit_time=now(),score=?",
			studentId, uint(questionType), lessonId, questionId, record.ID, score, score).Scan(&a)
	}
}

func (p PracticeService) UpdatePracticeAnswer(questionType questionType.QuestionType, questionId, lessonId, studentId uint, score uint) {
	global.GVA_DB.Raw("INSERT INTO tea_practice_answer(student_id,question_type, question_id, lesson_id,score)\nVALUES ( ?, ?, ?,?,?)\nON DUPLICATE KEY UPDATE score = GREATEST(score,?)", studentId, questionType, questionId, lessonId, score, score).Scan(nil)
	return
}

func (p PracticeService) FindHistoryAnswer(questionType questionType.QuestionType, ids []uint, studentId uint) *teachplanResp.History {
	histories := []teachplan.PracticeAnswer{}
	global.GVA_DB.Where("student_id=? and question_type=? and question_id in ?", studentId, questionType, ids).Find(&histories)
	history := teachplanResp.History{History: map[uint]*teachplanResp.HistoryItem{}}
	for _, t := range histories {
		item := teachplanResp.HistoryItem{
			Exist: true,
			Score: t.Score,
		}
		history.History[t.QuestionId] = &item
	}
	for _, id := range ids {
		if history.History[id] == nil {
			item := teachplanResp.HistoryItem{
				Exist: false,
				Score: 0,
			}
			history.History[id] = &item
		}
	}
	return &history
}

func (p PracticeService) CanNewPracticeRecord(lessonId uint, studentId uint) bool {
	str := fmt.Sprintf("newPracticeRecord:%d:%d", studentId, lessonId)
	_, err := global.GVA_REDIS.Get(context.Background(), str).Result()
	if err == nil {
		return false
	} else {
		global.GVA_REDIS.Set(context.Background(), str, true, 20*time.Minute)
		return true
	}
}
