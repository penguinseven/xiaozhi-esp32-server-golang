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

	// 无序列表标记 - 或 * 开头的行（注意：不是斜体包裹的星号）
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
)

// StripMarkdown 移除文本中的常见 Markdown 语法标记，保留纯文本内容。
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
func StripMarkdown(s string) string {
	if s == "" {
		return ""
	}

	// 1. 先处理跨行结构
	result := codeBlockRe.ReplaceAllString(s, "")

	// 2. 处理行内图片和链接
	result = imageRe.ReplaceAllString(result, "")      // 图片：直接移除
	result = linkRe.ReplaceAllString(result, "$1")      // 链接：保留显示文本

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

	// 5. 表格
	result = tableSeparatorRe.ReplaceAllString(result, "")
	result = tableRowRe.ReplaceAllString(result, "")

	// 6. 清理多余空行
	result = multiBlankRe.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
