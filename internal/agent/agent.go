package agent

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/google/uuid"

	"local-llm-lab/internal/db"
	"local-llm-lab/internal/llm"
	"local-llm-lab/internal/tools"
)

type Agent struct {
	llm llm.Client
	tools *tools.Registry
	store *db.Store
}

func New(client llm.Client, registry *tools.Registry, store *db.Store)*Agent{
	return &Agent{llm:client,tools:registry,store:store}
}

func(a *Agent)Run(ctx context.Context,input string)(string,string,error){
	runID:=uuid.NewString()
	if err:=a.store.CreateRun(runID,input);err!=nil{return "",runID,err}

	messages:=[]openai.ChatCompletionMessage{
		{Role:openai.ChatMessageRoleSystem,Content:"你是一个本地 Agent Infra 学习助手。需要读取工作区文件时使用工具；不要虚构工具结果。"},
		{Role:openai.ChatMessageRoleUser,Content:input},
	}

	specs:=make([]llm.ToolSpec,0)
	for _,t:=range a.tools.Specs(){
		specs=append(specs,llm.ToolSpec{Name:t.Name,Description:t.Description,Parameters:t.Parameters})
	}

	for stepNo:=1;stepNo<=8;stepNo++{
		stepID:=uuid.NewString()
		if err:=a.store.CreateStep(stepID,runID,stepNo,"LLM",lastUserInput(messages));err!=nil{
			return a.fail(runID,err)
		}

		res,err:=a.llm.Chat(ctx,messages,specs)
		if err!=nil{
			_ = a.store.CompleteStep(stepID,"FAILED","",err.Error())
			return a.fail(runID,err)
		}

		if len(res.ToolCalls)==0{
			_ = a.store.CompleteStep(stepID,"SUCCEEDED",res.Content,"")
			_ = a.store.CompleteRun(runID,"SUCCEEDED",res.Content,"")
			return res.Content,runID,nil
		}

		messages=append(messages,openai.ChatCompletionMessage{
			Role:openai.ChatMessageRoleAssistant,Content:res.Content,ToolCalls:res.ToolCalls,
		})

		for _,call:=range res.ToolCalls{
			toolStepID:=uuid.NewString()
			_ = a.store.CreateStep(toolStepID,runID,stepNo,"TOOL",call.Function.Arguments)

			callID:=call.ID
			if callID==""{callID=uuid.NewString()}
			if err:=a.store.CreateToolCall(callID,runID,toolStepID,call.Function.Name,call.Function.Arguments);err!=nil{
				return a.fail(runID,err)
			}

			tool,ok:=a.tools.Get(call.Function.Name)
			if !ok{
				errMsg:=fmt.Sprintf("unknown tool: %s",call.Function.Name)
				_ = a.store.CompleteToolCall(callID,"FAILED","",errMsg)
				_ = a.store.CompleteStep(toolStepID,"FAILED","",errMsg)
				messages=append(messages,openai.ChatCompletionMessage{Role:openai.ChatMessageRoleTool,ToolCallID:callID,Content:errMsg})
				continue
			}

			if err:=tools.ValidateJSON([]byte(call.Function.Arguments));err!=nil{
				_ = a.store.CompleteToolCall(callID,"FAILED","",err.Error())
				_ = a.store.CompleteStep(toolStepID,"FAILED","",err.Error())
				messages=append(messages,openai.ChatCompletionMessage{Role:openai.ChatMessageRoleTool,ToolCallID:callID,Content:err.Error()})
				continue
			}

			result,toolErr:=tool.Execute(ctx,[]byte(call.Function.Arguments))
			if toolErr!=nil{
				msg:="tool error: "+toolErr.Error()
				_ = a.store.CompleteToolCall(callID,"FAILED","",msg)
				_ = a.store.CompleteStep(toolStepID,"FAILED","",msg)
				messages=append(messages,openai.ChatCompletionMessage{Role:openai.ChatMessageRoleTool,ToolCallID:callID,Content:msg})
				continue
			}

			_ = a.store.CompleteToolCall(callID,"SUCCEEDED",result,"")
			_ = a.store.CompleteStep(toolStepID,"SUCCEEDED",result,"")
			messages=append(messages,openai.ChatCompletionMessage{Role:openai.ChatMessageRoleTool,ToolCallID:callID,Content:result})
		}

		_ = a.store.CompleteStep(stepID,"SUCCEEDED",res.Content,"")
	}

	return a.fail(runID,fmt.Errorf("maximum agent steps exceeded"))
}

func(a *Agent)fail(runID string,err error)(string,string,error){
	_ = a.store.CompleteRun(runID,"FAILED","",err.Error())
	return "",runID,err
}

func lastUserInput(ms []openai.ChatCompletionMessage)string{
	var b strings.Builder
	for _,m:=range ms{
		if m.Role==openai.ChatMessageRoleUser{
			if b.Len()>0{b.WriteString("\n")}
			b.WriteString(m.Content)
		}
	}
	return b.String()
}
