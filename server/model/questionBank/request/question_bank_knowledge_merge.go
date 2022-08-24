package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/questionBank"
)

type QuestionBankKnowledgeMergeSearch struct {
	questionBank.KnowledgeMerge
	request.PageInfo
}
