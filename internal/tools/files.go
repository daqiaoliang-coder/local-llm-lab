package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type listFilesTool struct{ root string }
func NewListFilesTool(root string) Tool { return &listFilesTool{root} }
func (t *listFilesTool) Name() string { return "list_files" }
func (t *listFilesTool) Description() string { return "列出工作区目录下的文件和子目录。" }
func (t *listFilesTool) Parameters() map[string]any {
	return map[string]any{"type":"object","properties":map[string]any{
		"path":map[string]any{"type":"string","description":"工作区内相对目录，例如 . 或 docs"},
	}}
}
func (t *listFilesTool) Execute(ctx context.Context, raw json.RawMessage) (string,error) {
	var in struct{ Path string `json:"path"` }
	if err:=json.Unmarshal(raw,&in); err!=nil{return "",err}
	if in.Path==""{in.Path="."}
	target,err:=safePath(t.root,in.Path);if err!=nil{return "",err}
	entries,err:=os.ReadDir(target);if err!=nil{return "",err}
	var b strings.Builder
	for _,e:=range entries {
		kind:="file";if e.IsDir(){kind="dir"}
		fmt.Fprintf(&b,"%s\t%s\n",kind,e.Name())
	}
	return b.String(),nil
}

type readFileTool struct{ root string }
func NewReadFileTool(root string) Tool { return &readFileTool{root} }
func (t *readFileTool) Name() string { return "read_file" }
func (t *readFileTool) Description() string { return "读取工作区内的文本文件。" }
func (t *readFileTool) Parameters() map[string]any {
	return map[string]any{"type":"object","properties":map[string]any{
		"path":map[string]any{"type":"string","description":"工作区内相对文件路径，例如 README.md"},
	},"required":[]string{"path"}}
}
func (t *readFileTool) Execute(ctx context.Context, raw json.RawMessage) (string,error) {
	var in struct{ Path string `json:"path"` }
	if err:=json.Unmarshal(raw,&in);err!=nil{return "",err}
	if in.Path==""{return "",fmt.Errorf("path is required")}
	target,err:=safePath(t.root,in.Path);if err!=nil{return "",err}
	data,err:=os.ReadFile(target);if err!=nil{return "",err}
	if len(data)>64*1024{data=data[:64*1024]}
	return string(data),nil
}

func safePath(root, requested string)(string,error){
	rootAbs,err:=filepath.Abs(root);if err!=nil{return "",err}
	targetAbs,err:=filepath.Abs(filepath.Join(rootAbs,requested));if err!=nil{return "",err}
	rel,err:=filepath.Rel(rootAbs,targetAbs);if err!=nil{return "",err}
	if rel==".." || strings.HasPrefix(rel,".."+string(os.PathSeparator)){
		return "",fmt.Errorf("path escapes workspace: %s",requested)
	}
	return targetAbs,nil
}
