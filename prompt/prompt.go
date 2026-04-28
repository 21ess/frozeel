package prompt

import (
	"fmt"
	"strings"

	"github.com/21ess/frozeel/provider"
)

// GameInfo 持有游戏所有的子信息
type GameInfo struct {
	BuildDesc string // build form description (optional)
}

// BuildSystemPrompt 为游戏的 GM Agent 组装 prompt（4 layers）
//
// Layers:
//  1. globalMem  — Memory.md 长期记忆（决定游戏的玩法）
//  2. roomMem    — Room.md 长期记忆 per-chat memory (每个房间的用户偏好)
//  3. gameInfo   — GameInfo 短期记忆部分（本局游戏的用户设置相关）
//  4. answer     — [CONFIDENTIAL] character card + reply format
func BuildSystemPrompt(globalMem, roomMem string, info GameInfo, answer *provider.Character, subject *provider.Subject) string {
	var sb strings.Builder

	// Layer 1: global memory
	sb.WriteString("=== 全局规则 ===\n")
	sb.WriteString(globalMem)
	sb.WriteString("\n\n")

	// Layer 2: room memory
	if roomMem != "" {
		sb.WriteString("=== 群聊记忆 ===\n")
		sb.WriteString(roomMem)
		sb.WriteString("\n\n")
	}

	// Layer 3: game session info
	if info.BuildDesc != "" {
		sb.WriteString("=== 本局信息 ===\n")
		sb.WriteString(fmt.Sprintf("题目描述：%s\n\n", info.BuildDesc))
	}

	// Layer 4: confidential answer card
	sb.WriteString("=== [机密] 本局答案 ===\n")
	sb.WriteString("以下信息严格保密，绝不能直接告诉玩家。\n\n")
	sb.WriteString(fmt.Sprintf("角色名: %s\n", answer.Name))
	if answer.Gender != "" {
		sb.WriteString(fmt.Sprintf("性别: %s\n", answer.Gender))
	}
	if answer.Birthday != "" {
		sb.WriteString(fmt.Sprintf("生日: %s\n", answer.Birthday))
	}
	if answer.Summary != "" {
		sb.WriteString(fmt.Sprintf("简介: %s\n", answer.Summary))
	}
	if len(answer.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(answer.Tags, ", ")))
	}
	if answer.Actor != "" {
		sb.WriteString(fmt.Sprintf("声优: %s\n", answer.Actor))
	}
	if subject.Name != "" {
		sb.WriteString(fmt.Sprintf("作品: %s\n", subject.Name))
	}

	sb.WriteString("\n=== 回复格式要求 ===\n")
	sb.WriteString("每次回复必须严格包含两部分：\n")
	sb.WriteString("1. 一个 JSON 代码块，包含 outcome 和 confidence：\n")
	sb.WriteString("```json\n{\"outcome\": \"correct\", \"confidence\": 0.95}\n```\n")
	sb.WriteString("outcome 取值: correct / wrong / close\n")
	sb.WriteString("confidence 取值: 0.0 ~ 1.0\n\n")
	sb.WriteString("2. 紧接着的自然语言回复，直接发送给玩家。\n")

	return sb.String()
}

// BuildHintPrompt creates a user message asking the agent to generate a hint.
func BuildHintPrompt(hintLevel int) string {
	return fmt.Sprintf(
		"请根据当前答案角色生成第 %d 级提示。提示应逐步透露更多信息，但不能直接说出角色名。"+
			"只需要回复提示内容，不需要 JSON 块。", hintLevel)
}

// BuildSummaryPrompt creates a system-level instruction for generating Room.md updates.
func BuildSummaryPrompt(currentRoomMd string) string {
	var sb strings.Builder
	sb.WriteString("请根据本局游戏的对话历史，生成或更新这个群聊的记忆文件。\n\n")

	if currentRoomMd != "" {
		sb.WriteString("当前群聊记忆内容：\n")
		sb.WriteString("```\n")
		sb.WriteString(currentRoomMd)
		sb.WriteString("\n```\n\n")
		sb.WriteString("请在此基础上更新，保留有价值的历史信息，添加本局新信息。\n\n")
	} else {
		sb.WriteString("这是该群聊的第一局游戏，请创建初始记忆。\n\n")
	}

	sb.WriteString("记忆文件应包含：\n")
	sb.WriteString("- 群聊语言偏好\n")
	sb.WriteString("- 活跃玩家档案（昵称、猜测风格、擅长领域）\n")
	sb.WriteString("- 历史游戏摘要（答案、获胜者、有趣时刻）\n")
	sb.WriteString("- 群聊氛围和特殊偏好\n\n")
	sb.WriteString("直接输出 markdown 内容，不需要代码围栏。")

	return sb.String()
}
