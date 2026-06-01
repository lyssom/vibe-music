# Vibe Echo - LLM 驱动多轮创作模式设计

## 概述

将多轮歌曲创作模式从硬编码状态机重构为 LLM 真正驱动的对话式交互。用户可以用自然语言与 AI 音乐助手对话，简单需求一次性完成，复杂需求逐步探索。

## 设计原则

1. **LLM 主导** - 让 LLM 决定对话流程，系统负责解析和执行
2. **混合模式** - 简单需求高效完成，复杂需求深度探索
3. **实时反馈** - 生成过程可视化，进度可见、可中断
4. **可迭代** - 生成后可随时修改任意部分

## 整体流程

### 简单模式（默认）

用户直接描述需求，LLM 判断可直接生成：

```
用户: "来首流行歌"
   ↓
LLM 回复 + JSON: {action: "generate", structure: [...]}
   ↓
系统解析 → 生成 → 显示进度 → 可播放
```

### 复杂模式（检测到需确认）

用户描述较复杂或有歧义，LLM 发起多轮对话：

```
用户: "我要写一首有爆发力副歌的流行歌"
   ↓
LLM 回复 + JSON: {action: "question", message: "...", options: [...]}
   ↓
用户回答 → 继续对话 → 最终 generate → 生成
```

## LLM 结构化输出格式

### 1. Action = "question"（等待用户回答）

```json
{
  "type": "question",
  "action": "question",
  "message": "你希望歌曲大概多长？",
  "options": ["2分钟", "3分钟", "4分钟以上"],
  "skip_asking": false
}
```

### 2. Action = "generate"（开始生成）

```json
{
  "type": "generate",
  "action": "generate",
  "structure": [
    {"id": "intro", "name": "前奏", "bars": 4},
    {"id": "verse", "name": "主歌", "bars": 8},
    {"id": "chorus", "name": "副歌", "bars": 8}
  ],
  "bpm": 120,
  "notes": "副歌部分力度加强，使用更强的鼓点"
}
```

### 3. Action = "done"（完成）

```json
{
  "type": "done",
  "action": "done",
  "message": "歌曲生成完成！"
}
```

## TUI 界面设计

### 对话模式

```
┌──────────────────────────────────────────────────────────────┐
│ ▓ SONG COMPOSER ▓                                           │
├──────────────────────────────────────────────────────────────┤
│ AI: 好的！你想要一首什么样的歌？                              │
│                                                              │
│ 你: 要一首流行歌，副歌要有爆发力                              │
│                                                              │
│ AI: 明白了！让我确认几个细节：                                │
│     你希望歌曲大概多长？                                      │
│     1. 2分钟左右   2. 3分钟左右   3. 4分钟以上                │
│                                                              │
│ 你: ___________                                             │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│ [1] 继续对话  ·  [2] 直接生成  ·  /quit 退出                  │
└──────────────────────────────────────────────────────────────┘
```

### 生成模式

```
┌──────────────────────────────────────────────────────────────┐
│ ▓ GENERATING ▓                                    [100%]    │
├──────────────────────────────────────────────────────────────┤
│ ┌─ 歌曲结构 ─────────────────────────────┐                  │
│ │ ▓▓▓▓▓▓▓▓▓░░ 75%                        │                  │
│ │ 前奏 ✓  主歌 ✓  副歌 ▓▓▓  桥段 ─  结尾 ─ │                  │
│ └─────────────────────────────────────────┘                  │
│                                                              │
│ [副歌] sound("bd sd").fast(2)                               │
│         chord("c3 e3 g3").gain(0.8)                          │
│         note("c4 e4 g4").delay(0.25)                        │
│         ↑ 正在生成...                                        │
├──────────────────────────────────────────────────────────────┤
│ [space] 播放当前进度  ·  [enter] 暂停生成  ·  [q] 放弃       │
└──────────────────────────────────────────────────────────────┘
```

## 核心接口设计

### Generator 接口扩展

```go
// StructuredResponse 是 LLM 返回的结构化响应
type StructuredResponse struct {
    Action  string   // "question" | "generate" | "done"
    Message string   // 回复文本
    Options []string // 问题选项（action=question 时）
    
    // 生成参数（action=generate 时）
    Structure []SectionSpec
    BPM       int
    Notes     string
}

type SectionSpec struct {
    ID    string
    Name  string
    Bars  int
}

type Generator interface {
    // GenerateWithStructuredResponse 返回结构化响应
    GenerateWithStructuredResponse(ctx context.Context, prompt string, history []Message) (*StructuredResponse, error)
    
    // GenerateSection 生成单段 DSL 代码
    GenerateSection(ctx context.Context, section SectionSpec, elements SongElements) (string, error)
}
```

### Composer 重构

移除硬编码的状态机，改为：

```go
type Composer struct {
    gen     Generator
    history []Message
    current *StructuredResponse
    song    *Song
}

// 对话循环
func (c *Composer) Chat(ctx context.Context, userInput string) (*StructuredResponse, error) {
    // 1. 追加用户输入到历史
    c.history = append(c.history, Message{Role: "user", Content: userInput})
    
    // 2. 调用 LLM 获取结构化响应
    resp, err := c.gen.GenerateWithStructuredResponse(ctx, "", c.history)
    if err != nil {
        return nil, err
    }
    
    // 3. 保存助手回复到历史
    c.history = append(c.history, Message{Role: "assistant", Content: resp.Message})
    
    // 4. 根据 action 处理
    c.current = resp
    
    switch resp.Action {
    case "question":
        return resp, nil  // 返回问题给用户
    case "generate":
        return resp, c.buildSong(ctx, resp)
    case "done":
        return resp, nil
    }
    
    return resp, nil
}
```

## System Prompt 设计

```
你是 Vibe Echo 的 AI 音乐创作助手。用户正在通过 DSL 语言创作音乐。

你的职责：
1. 理解用户的音乐需求（风格、情绪、节奏、乐器等）
2. 引导用户完善需求（如果信息不足）
3. 确认后生成歌曲结构并创建 DSL 代码

输出格式要求：
- 始终以 JSON 格式返回结构化响应
- JSON 必须是单行，放在 ```json ... ``` 代码块中
- 紧随 JSON 之后用自然语言回复用户

响应类型：
1. question - 当需要更多信息时，返回问题供用户选择
2. generate - 当信息充足时，返回歌曲结构开始生成
3. done - 生成完成

歌曲结构规范：
- 前奏 (intro): 2-8 小节
- 主歌 (verse): 8-16 小节
- 副歌 (chorus): 8-16 小节
- 桥段 (bridge): 4-8 小节
- 尾奏 (outro): 2-8 小节

常见结构示例：
- 简单流行: intro(4) → verse(8) → chorus(8) → verse(8) → chorus(8) → outro(4)
- 标准结构: intro(4) → verse(8) → pre-chorus(4) → chorus(8) → bridge(4) → chorus(8) → outro(4)
- 复杂结构: intro(4) → verse(8) → pre-chorus(4) → chorus(8) → verse(8) → chorus(8) → bridge(8) → chorus(16) → outro(4)

用户可以用中文或英文描述需求。
```

## 实现步骤

### Phase 1: 接口重构
1. 定义 StructuredResponse 类型
2. 扩展 Generator 接口
3. 重构 Composer，移除硬编码状态机

### Phase 2: LLM 集成
1. 设计 system prompt
2. 实现结构化输出解析
3. 测试各类型响应

### Phase 3: TUI 更新
1. 新的对话视图
2. 生成进度可视化
3. 流式代码展示

### Phase 4: 测试与迭代
1. 简单场景测试
2. 复杂场景测试
3. 用户体验优化

## 成功标准

1. 用户可以用自然语言完成一首歌曲的创作
2. 简单需求（1-2 句话）能在 3 步内完成
3. 复杂需求可以逐步探索
4. 生成过程可见、可中断、可修改
5. 生成后可随时修改任意部分