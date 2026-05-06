// Package tools
// Bangumi API 调用工具能力类
package tools

import (
	"context"

	"github.com/21ess/frozeel/provider/bangumi"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type BangumiSearchTool struct {
	bmProvider *bangumi.BmProvider
}

func NewBangumiSearchTool(bmProvider *bangumi.BmProvider) *BangumiSearchTool {
	return &BangumiSearchTool{
		bmProvider: bmProvider,
	}
}

func (b *BangumiSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_bangumi_character",
		Desc: "搜索 Bangumi acg 角色信息，包括角色名称、简介、性别、生日等",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (b *BangumiSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	//TODO implement me
	panic("implement me")
}
