package text

import (
	"regexp"
	"strings"
)

var (
	// 代码块 ``` 或 ~~~
	codeBlockRe = regexp.MustCompile("(?s)```[^\\n]*\\n.*?```|(?s)~~~[^\\n]*\\n.*?~~~")

	// 行内代码 `code`
	inlineCodeRe = regexp.MustCompile("`[^`]+`")

	// 图片 ![alt](url)
	imageRe = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)

	// 链接 [text](url)
	linkRe = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)

	// 引用标记 > 开头的行
	blockquoteRe = regexp.MustCompile(`(?m)^>\s?`)

	// 无序列表标记 - 或 * 开头的行
	unorderedListRe = regexp.MustCompile(`(?m)^[*-]\s+`)

	// 有序列表标记 1. 2. 等
	orderedListRe = regexp.MustCompile(`(?m)^\d+\.\s+`)

	// 水平分隔线 --- *** ___
	hrRe = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)

	// 标题标记 # 开头
	headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+`)

	// 加粗 **text** 或 __text__
	boldRe = regexp.MustCompile(`\*{2}(.+?)\*{2}|_{2}(.+?)_{2}`)

	// 斜体 *text* 或 _text_（单独一行，非列表标记）
	italicRe = regexp.MustCompile(`\*([^*\s][^*]*?[^*\s])\*|_([^_\s][^_]*?[^_\s])_`)

	// 删除线 ~~text~~
	strikethroughRe = regexp.MustCompile(`~~(.+?)~~`)

	// 表格分隔行 |---|---|
	tableSeparatorRe = regexp.MustCompile(`(?m)^\|?[-:|]+\|[-:|]+\|?$`)

	// 表格行 | col1 | col2 |
	tableRowRe = regexp.MustCompile(`(?m)^\|.+\|\s*$`)

	// 多个连续空行
	multiBlankRe = regexp.MustCompile(`\n{3,}`)

	// 双空格
	doubleSpaceRe = regexp.MustCompile(`  +`)

	// 行尾空格
	trailingSpaceRe = regexp.MustCompile(` +\n`)

	// 行首空格
	leadingSpaceRe = regexp.MustCompile(`\n +`)

	// 表情符号 - 常见表情符号 Unicode 范围
	emojiRe = regexp.MustCompile(
		`[\x{1F600}-\x{1F64F}` +
			`\x{1F300}-\x{1F5FF}` +
			`\x{1F680}-\x{1F6FF}` +
			`\x{1F900}-\x{1F9FF}` +
			`\x{1FA70}-\x{1FAFF}` +
			`\x{2600}-\x{27BF}` +
			`\x{FE00}-\x{FE0F}` +
			`\x{200D}` +
			`\x{2139}` +
			`\x{231A}\x{231B}\x{2328}\x{23CF}` +
			`\x{23E9}-\x{23F3}\x{23F8}-\x{23FA}` +
			`\x{24C2}` +
			`\x{25AA}\x{25AB}\x{25B6}\x{25C0}\x{25FB}-\x{25FE}` +
			`\x{2600}-\x{2604}\x{260E}\x{2611}\x{2614}\x{2615}\x{2618}\x{261D}` +
			`\x{2620}\x{2622}\x{2623}\x{2626}\x{262A}\x{262E}\x{262F}\x{2638}-\x{263A}` +
			`\x{2640}\x{2642}\x{2648}-\x{2653}\x{265F}\x{2660}\x{2663}\x{2665}\x{2666}\x{2668}` +
			`\x{267B}\x{267E}\x{267F}\x{2692}-\x{2697}\x{2699}\x{269B}\x{269C}` +
			`\x{26A0}\x{26A1}\x{26AA}\x{26AB}\x{26B0}\x{26B1}\x{26BD}\x{26BE}` +
			`\x{26C4}\x{26C5}\x{26C8}\x{26CE}\x{26CF}\x{26D1}\x{26D3}\x{26D4}` +
			`\x{26E9}\x{26EA}\x{26F0}-\x{26F5}\x{26F7}-\x{26FA}\x{26FD}` +
			`\x{2702}\x{2705}\x{2708}-\x{270D}\x{270F}\x{2712}\x{2714}\x{2716}\x{271D}\x{2721}\x{2728}` +
			`\x{2733}\x{2734}\x{2744}\x{2747}\x{274C}\x{274E}\x{2753}-\x{2755}\x{2757}` +
			`\x{2763}\x{2764}\x{2795}-\x{2797}\x{27A1}\x{27B0}\x{27BF}` +
			`]`,
	)
)

// StripMarkdown 移除文本中的常见 Markdown 语法标记和表情符号，保留纯文本内容。
// 适用于 TTS 朗读前清洗 LLM 返回的 Markdown 格式文本。
//
// 处理以下标记：
//   - 代码块（``` 和 ~~~）
//   - 行内代码（`code`）
//   - 图片（![alt](url)）
//   - 链接（[text](url) → text）
//   - 引用（> 开头的行）
//   - 无序列表（- 或 * 开头的行）
//   - 有序列表（1. 2. 等）
//   - 水平分隔线（---, ***, ___）
//   - 标题（# 开头）
//   - 加粗（**text**）
//   - 斜体（*text*）
//   - 删除线（~~text~~）
//   - 表格
//   - 表情符号（😊 ❤️ 👍 等）
func StripMarkdown(s string) string {
	if s == "" {
		return ""
	}

	// 1. 先处理跨行结构
	result := codeBlockRe.ReplaceAllString(s, "")

	// 2. 处理行内图片和链接
	result = imageRe.ReplaceAllString(result, "")       // 图片：直接移除
	result = linkRe.ReplaceAllString(result, "$1")       // 链接：保留显示文本

	// 3. 处理行级标记
	result = hrRe.ReplaceAllString(result, "")
	result = headingRe.ReplaceAllString(result, "")
	result = blockquoteRe.ReplaceAllString(result, "")
	result = unorderedListRe.ReplaceAllString(result, "")
	result = orderedListRe.ReplaceAllString(result, "")

	// 4. 处理内联标记
	result = boldRe.ReplaceAllString(result, "$1$2")
	result = italicRe.ReplaceAllString(result, "$1$2")
	result = strikethroughRe.ReplaceAllString(result, "$1")
	result = inlineCodeRe.ReplaceAllString(result, "")

	// 5. 移除表情符号
	result = emojiRe.ReplaceAllString(result, "")

	// 6. 表格
	result = tableSeparatorRe.ReplaceAllString(result, "")
	result = tableRowRe.ReplaceAllString(result, "")

	// 7. 清理多余空白
	result = multiBlankRe.ReplaceAllString(result, "\n\n")
	result = doubleSpaceRe.ReplaceAllString(result, " ")
	result = trailingSpaceRe.ReplaceAllString(result, "\n")
	result = leadingSpaceRe.ReplaceAllString(result, "\n")

	return strings.TrimSpace(result)
}
