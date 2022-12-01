package questionBank

import (
	"github.com/gin-gonic/gin"
	"github.com/prl26/exam-system/server/model/common/response"
	ojReq "github.com/prl26/exam-system/server/model/oj/request"
	ojResp "github.com/prl26/exam-system/server/model/oj/response"
	questionBankBo "github.com/prl26/exam-system/server/model/questionBank/bo"
	questionBankEnum "github.com/prl26/exam-system/server/model/questionBank/enum"
	"github.com/prl26/exam-system/server/service"
	"github.com/prl26/exam-system/server/utils"
)

//
//import (
//	"github.com/gin-gonic/gin"
//	"github.com/prl26/exam-system/server/model/common/response"
//	ojReq "github.com/prl26/exam-system/server/model/oj/request"
//	ojResp "github.com/prl26/exam-system/server/model/oj/response"
//	"github.com/prl26/exam-system/server/service"
//	"github.com/prl26/exam-system/server/utils"
//)
//
type OjApi struct {
}

//
var (
	judgeService          = service.ServiceGroupApp.OjServiceServiceGroup.JudgeService
	cLanguageService      = &service.ServiceGroupApp.OjServiceServiceGroup.CLanguageService
	commonService         = &service.ServiceGroupApp.OjServiceServiceGroup.CommonService
	supplyBlankService    = service.ServiceGroupApp.OjServiceServiceGroup.SupplyBlankService
	multipleChoiceService = service.ServiceGroupApp.OjServiceServiceGroup.MultipleChoiceService
)

func (*OjApi) CheckJudge(c *gin.Context) {
	var r ojReq.CheckJudge
	_ = c.ShouldBindJSON(&r)
	verify := utils.Rules{
		"Id": {utils.NotEmpty()},
	}
	if err := utils.Verify(r, verify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := judgeService.Check(r.Id, r.Answer)
	if err != nil {
		response.FailWithMessage("找不到该题目", c)
		return
	}
	response.OkWithData(result, c)
	return
}

func (*OjApi) CheckProgram(c *gin.Context) {
	var r ojReq.CheckProgramm
	_ = c.ShouldBindJSON(&r)
	verify := utils.Rules{
		"Id":         {utils.NotEmpty()},
		"Code":       {utils.NotEmpty()},
		"LanguageId": {utils.NotEmpty()},
	}
	if err := utils.Verify(r, verify); err != nil {
		ojResp.ErrorHandle(c, err)
		return
	}
	program, err := commonService.FindProgram(r.Id)
	if err != nil {
		ojResp.ErrorHandle(c, err)
		return
	}
	support := questionBankBo.LanguageSupport{}
	err = support.Deserialize(program.LanguageSupports, questionBankEnum.LanguageType(r.LanguageId))
	if err != nil {
		ojResp.ErrorHandle(c, err)
		return
	}
	cases := questionBankBo.ProgramCases{}
	err = cases.Deserialize(program.ProgramCases)
	if err != nil {
		ojResp.ErrorHandle(c, err)
		return
	}
	switch r.LanguageId {
	case questionBankEnum.C_LANGUAGE:
		result, err := cLanguageService.Check(r.Code, support.LanguageLimit, cases)
		if err != nil {
			ojResp.ErrorHandle(c, err)
			return
		}
		response.OkWithData(result, c)
		return
	default:
		ojResp.ErrorHandle(c, err)
		return
	}
}

func (*OjApi) CheckSupplyBlank(c *gin.Context) {
	var r ojReq.CheckSupplyBlank
	_ = c.ShouldBindJSON(&r)
	verify := utils.Rules{
		"Id":      {utils.NotEmpty()},
		"Answers": {utils.NotEmpty()},
	}
	if err := utils.Verify(r, verify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, _, err := supplyBlankService.Check(r.Id, r.Answers)
	if err != nil {
		response.FailWithMessage("找不到该题目", c)
		return
	}
	response.OkWithData(result, c)
	return
}

func (*OjApi) CheckMultipleChoice(c *gin.Context) {
	var r ojReq.CheckMultipleChoice
	_ = c.ShouldBindJSON(&r)
	verify := utils.Rules{
		"Id":      {utils.NotEmpty()},
		"Answers": {utils.NotEmpty()},
	}
	if err := utils.Verify(r, verify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := multipleChoiceService.Check(r.Id, r.Answers)
	if err != nil {
		response.FailWithMessage("找不到该题目", c)
		return
	}
	response.OkWithData(result, c)
	return
}
