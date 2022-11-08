package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/prl26/exam-system/server/model/common/response"
	"strconv"
)

type LessonApi struct {

}

func (*LessonApi) FindLessonDetail(c*gin.Context) {
	query := c.Query("id")
	idInt, err := strconv.Atoi(query)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	idUint := uint(idInt)
	detail,err:= lessonService.FindLessonDetail(idUint)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}
