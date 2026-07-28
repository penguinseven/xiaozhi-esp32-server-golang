package text

import (
	"testing"
)

func TestStripMarkdown_EmptyString(t *testing.T) {
	if got := StripMarkdown(""); got != "" {
		t.Errorf("StripMarkdown('') = %q, want ''", got)
	}
}

func TestStripMarkdown_PlainText(t *testing.T) {
	input := "你好，今天天气不错。"
	if got := StripMarkdown(input); got != input {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, input)
	}
}

func TestStripMarkdown_Bold(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"双星号加粗", "这是一段**重要**的文字", "这是一段重要的文字"},
		{"双下划线加粗", "这是一段__重要__的文字", "这是一段重要的文字"},
		{"跨词加粗", "**非常非常**重要", "非常非常重要"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_Italic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"单星号斜体", "这是一段*斜体*文字", "这是一段斜体文字"},
		{"单下划线斜体", "这是一段_斜体_文字", "这是一段斜体文字"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_InlineCode(t *testing.T) {
	input := "请运行 `make build` 命令"
	want := "请运行 命令"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_CodeBlock(t *testing.T) {
	input := "代码如下：\n```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n请执行。"
	want := "代码如下：\n\n请执行。"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_TildeCodeBlock(t *testing.T) {
	input := "示例：\n~~~\ncode here\n~~~\n完毕。"
	want := "示例：\n\n完毕。"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Link(t *testing.T) {
	input := "访问 [OpenAI](https://openai.com) 官网"
	want := "访问 OpenAI 官网"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Image(t *testing.T) {
	input := "图片：![logo](https://example.com/logo.png)"
	want := "图片："
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Heading(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"一级标题", "# 标题一\n\n正文内容", "标题一\n\n正文内容"},
		{"二级标题", "## 标题二\n\n正文内容", "标题二\n\n正文内容"},
		{"六级标题", "###### 标题六\n\n正文内容", "标题六\n\n正文内容"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_Blockquote(t *testing.T) {
	input := "> 这是一段引用\n> 这是第二行\n\n正文"
	want := "这是一段引用\n这是第二行\n\n正文"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_UnorderedList(t *testing.T) {
	input := "物品清单：\n- 苹果\n- 香蕉\n- 橙子"
	want := "物品清单：\n苹果\n香蕉\n橙子"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_OrderedList(t *testing.T) {
	input := "步骤：\n1. 打开电源\n2. 按下按钮\n3. 等待完成"
	want := "步骤：\n打开电源\n按下按钮\n等待完成"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_HorizontalRule(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"三个减号", "上面\n---\n下面", "上面\n\n下面"},
		{"三个星号", "上面\n***\n下面", "上面\n\n下面"},
		{"三个下划线", "上面\n___\n下面", "上面\n\n下面"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_Strikethrough(t *testing.T) {
	input := "这是~~删除~~的文字"
	want := "这是删除的文字"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Table(t *testing.T) {
	input := "| 姓名 | 年龄 |\n|------|------|\n| 张三 | 25 |\n| 李四 | 30 |"
	want := ""
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_TableWithText(t *testing.T) {
	input := "以下是表格：\n| 姓名 | 年龄 |\n|------|------|\n| 张三 | 25 |\n\n完毕。"
	want := "以下是表格：\n\n\n完毕。"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Emoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"笑脸", "眼角含笑 😊", "眼角含笑"},
		{"哭脸", "😂 笑死我了", "笑死我了"},
		{"爱心", "❤️ 爱心", "爱心"},
		{"点赞", "👍 点赞", "点赞"},
		{"庆祝", "🎉 庆祝", "庆祝"},
		{"混合标记", "**重要** 😊 完成", "重要 完成"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_NoFalsePositive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"数学乘法", "2 * 3 = 6", "2 * 3 = 6"},
		{"文件路径", "~/docs/README.md", "~/docs/README.md"},
		{"版本号", "v2.0.1", "v2.0.1"},
		{"IP地址", "192.168.1.1", "192.168.1.1"},
		{"C指针", "int *p = malloc(4)", "int *p = malloc(4)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkdown(tt.input); got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown_BoldItalicCombined(t *testing.T) {
	input := "***加粗斜体*** 文字"
	want := "加粗斜体 文字"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_MultipleBlankLines(t *testing.T) {
	input := "第一段\n\n\n\n第二段"
	want := "第一段\n\n第二段"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_LeadingTrailingSpaces(t *testing.T) {
	input := "  你好世界  "
	want := "你好世界"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(%q) = %q, want %q", input, got, want)
	}
}

func TestStripMarkdown_Complex(t *testing.T) {
	input := "# 使用说明 😊\n\n## 安装\n\n运行 `make build` 命令 ❤️\n\n> 注意：请确保版本 >= 2.0\n\n1. **启动**：检查端口 👍\n2. *检查*：网络设置\n\n- 选项A\n- 选项B\n\n更多信息：~/docs/README.md 🎉"
	want := "使用说明\n\n安装\n\n运行 命令\n\n注意：请确保版本 >= 2.0\n\n启动：检查端口\n检查：网络设置\n\n选项A\n选项B\n\n更多信息：~/docs/README.md"
	if got := StripMarkdown(input); got != want {
		t.Errorf("StripMarkdown(complex)\ngot:  %q\nwant: %q", got, want)
	}
}
