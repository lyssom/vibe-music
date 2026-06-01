package song

import "strings"

// Explorer manages the ten elements exploration flow
type Explorer struct {
	session *Session
}

// NewExplorer creates a new explorer
func NewExplorer(session *Session) *Explorer {
	return &Explorer{session: session}
}

// determinePriority returns element priorities based on current context
func (e *Explorer) determinePriority(turn int) []string {
	mode := e.session.GetState().Mode
	elements := e.session.GetElements()

	var priorities []string

	switch mode {
	case ModeSimple:
		return []string{"genre"}
	case ModeStandard:
		priorities = []string{"genre", "emotion", "rhythm", "instrumentation"}
	case ModeFull:
		if elements.Genre == "" {
			priorities = []string{"genre"}
		} else if elements.Emotion == "" {
			priorities = []string{"emotion"}
		} else if elements.Rhythm == "" {
			priorities = []string{"rhythm"}
		} else if elements.Instrumentation == "" {
			priorities = []string{"instrumentation"}
		} else if elements.Scale == "" {
			priorities = []string{"scale"}
		} else if elements.Harmony == "" {
			priorities = []string{"harmony"}
		} else if elements.BPM == 0 {
			priorities = []string{"bpm"}
		} else if elements.Dynamic == "" {
			priorities = []string{"dynamic"}
		} else if elements.Structure == "" {
			priorities = []string{"structure"}
		} else {
			priorities = []string{"techniques"}
		}
	}

	return priorities
}

// NextQuestion generates the next question based on turn
func (e *Explorer) NextQuestion(turn int) Question {
	priorities := e.determinePriority(turn)
	if len(priorities) == 0 {
		return Question{Kind: "done", Text: "我已经了解了你需要的音乐要素。"}
	}

	kind := priorities[0]

	switch kind {
	case "genre":
		return Question{
			Kind:    "genre",
			Text:    "你想要什么风格的音乐？（爵士、流行、摇滚、电子、古典、民谣、蓝调...）",
			Options: []string{"爵士 Jazz", "流行 Pop", "摇滚 Rock", "电子 Electronic", "古典 Classical", "蓝调 Blues", "民谣 Folk"},
		}
	case "emotion":
		return Question{
			Kind:    "emotion",
			Text:    "这首歌想表达什么情感或情绪？",
			Options: []string{"欢快 Happy", "忧伤 Melancholic", "浪漫 Romantic", "激昂 Energetic", "平静 Calm", "怀旧 Nostalgic", "神秘 Mysterious"},
		}
	case "rhythm":
		return Question{
			Kind:    "rhythm",
			Text:    "你偏好什么节奏感？",
			Options: []string{"快节奏 Upbeat", "慢节奏 Slow", "紧凑 Tight", "宽松 Laid-back", "强律动 Groovy", "柔和 Soft"},
		}
	case "instrumentation":
		return Question{
			Kind:    "instrumentation",
			Text:    "想要什么配器和音色？",
			Options: []string{"钢琴三重奏 Piano Trio", "电声乐队 Electric Band", "管弦乐团 Orchestra", "原声吉他 Acoustic", "合成器 Synth", "人声 Vocal"},
		}
	case "scale":
		return Question{
			Kind:    "scale",
			Text:    "喜欢什么调式和音阶？",
			Options: []string{"大调 Major", "小调 Minor", "蓝调音阶 Blues", "五声音阶 Pentatonic", "调式 Modal", "爵士旋律 Minor Jazz"},
		}
	case "harmony":
		return Question{
			Kind:    "harmony",
			Text:    "和声走向偏好？",
			Options: []string{"功能和弦 Functional", "爵士进行 Jazz Changes", "调式自由 Modal", "循环和声 Loop Harmony", "无和声 Atonal"},
		}
	case "bpm":
		return Question{
			Kind:    "bpm",
			Text:    "想要什么速度（BPM）？",
			Options: []string{"慢速 Slow (60-80)", "中速 Medium (80-110)", "快速 Fast (110-140)", "超快 Very Fast (140+)", "由你决定"},
		}
	case "dynamic":
		return Question{
			Kind:    "dynamic",
			Text:    "动态范围偏好？",
			Options: []string{"柔和-激昂 Soft-Loud", "平稳 Flat", "渐强渐弱 Crescendo", "持续中强 Constant Mid"},
		}
	case "structure":
		return Question{
			Kind:    "structure",
			Text:    "希望歌曲结构是怎样的？",
			Options: []string{"由你推荐", "简单 AABA", "主副结构 Verse-Chorus", "自由结构 Free Form"},
		}
	case "techniques":
		return Question{
			Kind:    "techniques",
			Text:    "想要什么特殊技巧？",
			Options: []string{"即兴 Solo", "和声转位 Chord Inversion", "变奏 Variation", "转调 Modulation", "无特别要求"},
		}
	default:
		return Question{Kind: "done", Text: "要素探索完成。"}
	}
}

// IsComplete checks if all required elements are filled
func (e *Explorer) IsComplete() bool {
	elements := e.session.GetElements()
	mode := e.session.GetState().Mode

	switch mode {
	case ModeSimple:
		return elements.Genre != ""
	case ModeStandard:
		return elements.Genre != "" && elements.Emotion != "" && elements.Rhythm != "" && elements.Instrumentation != ""
	case ModeFull:
		return elements.Genre != "" && elements.Emotion != "" && elements.Rhythm != "" &&
			elements.Instrumentation != "" && elements.Scale != "" && elements.Harmony != "" &&
			elements.BPM != 0 && elements.Dynamic != "" && elements.Structure != "" && elements.Techniques != ""
	}
	return false
}

// DetermineMode determines interaction mode from initial prompt
func DetermineMode(initialPrompt string) InteractionMode {
	// Simple patterns
	if strings.Contains(initialPrompt, "beat") || strings.Contains(initialPrompt, "loop") ||
		strings.Contains(initialPrompt, "pattern") || strings.Contains(initialPrompt, "drum") ||
		strings.Contains(initialPrompt, "节奏") || strings.Contains(initialPrompt, "循环") {
		return ModeSimple
	}

	// Full patterns (song-related keywords)
	if strings.Contains(initialPrompt, "song") || strings.Contains(initialPrompt, "主歌") ||
		strings.Contains(initialPrompt, "副歌") || strings.Contains(initialPrompt, "桥段") ||
		strings.Contains(initialPrompt, "album") || strings.Contains(initialPrompt, "track") {
		return ModeFull
	}

	return ModeStandard
}
