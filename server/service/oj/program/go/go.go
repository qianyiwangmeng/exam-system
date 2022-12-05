package goLanguage

import (
	"context"
	"fmt"
	"github.com/prl26/exam-system/server/global"
	"github.com/prl26/exam-system/server/model/oj"
	exception "github.com/prl26/exam-system/server/model/oj/error"
	ojResp "github.com/prl26/exam-system/server/model/oj/response"
	questionBankBo "github.com/prl26/exam-system/server/model/questionBank/bo"
	"github.com/prl26/exam-system/server/pb"
	"strconv"
	"strings"
	"time"
)

type GoLanguage struct {
	ExecutorClient                    pb.ExecutorClient
	GC_PATH                           string
	DEFAULT_COMPILE_CPU_TIME_LIMIT    uint64
	DEFAULT_COMPILE_MEMORY_TIME_LIMIT uint64
	DEFAULT_JUDGE_CPU_TIME_LIMI       uint64
	DEFAULT_JUDGE_MEMORY_LIMIT        uint64
}

const STDOUT = "stdout"
const STDERR = "stderr"

const DEFAULT_CODE_NAME string = "a.go"
const DEFAULT_FILE_NAME string = "a"

var replacer = strings.NewReplacer("\n", "", " ", "", "\t", "")

// 注意此处并不是 服务启动的真实值
// 此为服务启动的默认值   根据 config文件的配置 之后会进行依赖注入 来修改上面的值

const FILE_FAILED_DURATION time.Duration = 3 * time.Minute

func (c *GoLanguage) Check(code string, limit questionBankBo.LanguageLimit, cases questionBankBo.ProgramCases) ([]*ojResp.Submit, uint, error) {
	fileID, err := c.compile(code)
	if err != nil {
		return nil, 0, exception.CompileError{Msg: err.Error()}
	}
	go func() {
		after := time.After(FILE_FAILED_DURATION)
		<-after
		err := c.Delete(fileID)
		if err != nil {
			global.GVA_LOG.Error(err.Error())
			return
		}
	}()
	return c.Judge(fileID, limit, cases)
}

func (c *GoLanguage) Compile(code string) (string, *time.Time, error) {
	fileID, err := c.compile(code)
	if err != nil {
		return "", nil, exception.CompileError{Msg: err.Error()}
	}
	failedTime := time.Now().Add(FILE_FAILED_DURATION)
	go func() {
		after := time.After(FILE_FAILED_DURATION)
		<-after
		err := c.Delete(fileID)
		if err != nil {
			global.GVA_LOG.Error(err.Error())
			return
		}
	}()
	return fileID, &failedTime, nil
}

func (c *GoLanguage) compile(code string) (string, error) {
	input := &pb.Request_File{
		File: &pb.Request_File_Memory{
			Memory: &pb.Request_MemoryFile{
				Content: []byte(code)},
		},
	}
	stdio := &pb.Request_File_Memory{
		Memory: &pb.Request_MemoryFile{
			Content: []byte("")},
	}
	stout := &pb.Request_File_Pipe{
		Pipe: &pb.Request_PipeCollector{
			Name: STDOUT,
			Max:  10240},
	}
	stderr := &pb.Request_File_Pipe{
		Pipe: &pb.Request_PipeCollector{
			Name: STDERR,
			Max:  10240,
		},
	}
	cmd := &pb.Request_CmdType{
		Env:  []string{"GOCACHE=/tmp/go-build", "GOPATH=/tmp/gopath"},
		Args: []string{c.GC_PATH, "build", "-o", DEFAULT_FILE_NAME, DEFAULT_CODE_NAME},
		Files: []*pb.Request_File{
			{
				File: stdio,
			}, {
				File: stout,
			}, {
				File: stderr,
			},
		},
		CpuTimeLimit: c.DEFAULT_COMPILE_CPU_TIME_LIMIT,
		MemoryLimit:  c.DEFAULT_COMPILE_MEMORY_TIME_LIMIT,
		ProcLimit:    100,
		CopyIn: map[string]*pb.Request_File{
			DEFAULT_CODE_NAME: input,
		},
		CopyOut: []*pb.Request_CmdCopyOutFile{
			{
				Name: STDOUT,
			}, {
				Name: STDERR,
			},
		},
		CopyOutCached: []*pb.Request_CmdCopyOutFile{
			{
				Name: DEFAULT_FILE_NAME,
			},
		},
	}
	exec, err := c.ExecutorClient.Exec(context.Background(), &pb.Request{
		Cmd: []*pb.Request_CmdType{
			cmd,
		},
	})
	if err != nil {
		return "", err
	}
	result := exec.GetResults()[0]
	if result.Status != pb.Response_Result_Accepted {
		//说明出现了错误
		//此数应该打日志
		return "", fmt.Errorf("compile:%s", string(result.Files[STDERR]))
	}
	return exec.GetResults()[0].GetFileIDs()[DEFAULT_FILE_NAME], nil
}

func (c *GoLanguage) Delete(id string) error {
	_, err := c.ExecutorClient.FileDelete(context.Background(), &pb.FileID{FileID: id})
	if err != nil {
		return err
	}
	return nil
}

func (c *GoLanguage) Judge(fileId string, limit questionBankBo.LanguageLimit, cases questionBankBo.ProgramCases) ([]*ojResp.Submit, uint, error) {
	n := len(cases)
	submits := make([]*ojResp.Submit, n)
	cmds := make([]*pb.Request_CmdType, n)
	for i, programmCase := range cases {
		cmds[i] = c.makeCmd(fileId, programmCase.Input, limit)
	}
	exec, err := c.ExecutorClient.Exec(context.Background(), &pb.Request{
		Cmd: cmds,
	})
	if err != nil {
		return nil, 0, err
	}
	results := exec.GetResults()
	var sum uint
	for i, result := range results {
		submits[i] = &ojResp.Submit{Name: cases[i].Name, Score: 0, ExecuteSituation: oj.ExecuteSituation{
			ResultStatus: result.Status.String(), ExitStatus: int(result.ExitStatus), Time: uint(result.Time), Memory: uint(result.Memory), Runtime: uint(result.RunTime)},
		}
		if result.Status == pb.Response_Result_Accepted {
			standardAnswer := strings.ReplaceAll(string(result.Files[STDOUT]), "\r\n", "\n")
			actualAnswer := strings.ReplaceAll(cases[i].Output, "\r\n", "\n")
			if standardAnswer != actualAnswer {
				if replacer.Replace(standardAnswer) == replacer.Replace(actualAnswer) {
					result.Status = pb.Response_Result_PartiallyCorrect
				} else {
					result.Status = pb.Response_Result_WrongAnswer
				}
			} else {
				submits[i].Score = cases[i].Score
				sum += cases[i].Score
			}
		}
	}
	return submits, sum, nil
}

func (c *GoLanguage) Execute(fileId string, input string, programmLimit questionBankBo.LanguageLimit) (string, *oj.ExecuteSituation, error) {
	cmd := c.makeCmd(fileId, input, programmLimit)
	result, err := c.ExecutorClient.Exec(context.Background(), &pb.Request{
		Cmd: []*pb.Request_CmdType{
			cmd,
		},
	})
	if err != nil {
		return "", nil, err
	}
	response := result.Results[0]
	var out string
	var executeSituation = &oj.ExecuteSituation{ResultStatus: response.Status.String(), ExitStatus: int(response.ExitStatus), Time: uint(response.Time), Memory: uint(response.Memory), Runtime: uint(response.RunTime)}
	if response.Status == pb.Response_Result_Accepted {
		out = string(response.Files[STDOUT])
	}
	return out, executeSituation, nil
}

func (c *GoLanguage) makeCmd(fileId string, input string, programmLimit questionBankBo.LanguageLimit) *pb.Request_CmdType {
	inputFile := &pb.Request_File_Memory{
		Memory: &pb.Request_MemoryFile{
			Content: []byte(input),
		},
	}
	stout := &pb.Request_File_Pipe{
		Pipe: &pb.Request_PipeCollector{
			Name: STDOUT,
			Max:  10240},
	}
	stderr := &pb.Request_File_Pipe{
		Pipe: &pb.Request_PipeCollector{
			Name: STDERR,
			Max:  10240,
		},
	}
	cmd := &pb.Request_CmdType{
		Env:  []string{"PATH=/usr/local/bin:/usr/bin:/bin"},
		Args: []string{DEFAULT_FILE_NAME},
		Files: []*pb.Request_File{{
			File: inputFile,
		}, {
			File: stout,
		}, {
			File: stderr,
		},
		},
		ProcLimit: 50,
		CopyIn: map[string]*pb.Request_File{
			DEFAULT_FILE_NAME: {
				File: &pb.Request_File_Cached{
					Cached: &pb.Request_CachedFile{
						FileID: fileId,
					},
				},
			},
		},
		CopyOut: []*pb.Request_CmdCopyOutFile{
			{
				Name: STDOUT,
			}, {
				Name: STDERR,
			},
		},
	}

	cmd = c.cmdLimit(programmLimit, cmd)

	return cmd
}

func (c *GoLanguage) cmdLimit(programmLimit questionBankBo.LanguageLimit, cmd *pb.Request_CmdType) *pb.Request_CmdType {
	if programmLimit.CpuLimit != nil {
		cmd.CpuTimeLimit = uint64(*programmLimit.CpuLimit)
	} else {
		cmd.CpuTimeLimit = c.DEFAULT_COMPILE_CPU_TIME_LIMIT
	}
	if programmLimit.MemoryLimit != nil {
		cmd.MemoryLimit = uint64(*programmLimit.MemoryLimit)
	} else {
		cmd.MemoryLimit = c.DEFAULT_JUDGE_MEMORY_LIMIT
	}
	if programmLimit.ProcLimit != nil {
		cmd.ProcLimit = uint64(*programmLimit.ProcLimit)
	}
	if programmLimit.CpuSetLimit != nil {
		cmd.CpuSetLimit = strconv.Itoa(*programmLimit.CpuSetLimit)
	}
	if programmLimit.StackLimit != nil {
		cmd.StackLimit = uint64(*programmLimit.StackLimit)
	}
	if programmLimit.CpuRateLimit != nil {
		cmd.CpuRateLimit = uint64(*programmLimit.CpuRateLimit)
	}
	if programmLimit.ClockLimit != nil {
		cmd.ClockTimeLimit = uint64(*programmLimit.ClockLimit)
	}
	if programmLimit.StrictMemoryLimit != nil && *programmLimit.StackLimit == 1 {
		cmd.StrictMemoryLimit = true
	}
	return cmd
}
