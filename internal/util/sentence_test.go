package util

import "testing"

func TestExtractSmartSentencesKeepsTimeTogether(t *testing.T) {
	text := "根据系统时间，现在是2026年3月20日 星期五 02:37:04。"

	sentences, remaining := ExtractSmartSentences(text, 2, 100, false)

	if remaining != "" {
		t.Fatalf("expected no remaining text, got %q", remaining)
	}
	if len(sentences) != 1 {
		t.Fatalf("expected 1 sentence, got %d: %#v", len(sentences), sentences)
	}
	if sentences[0] != text {
		t.Fatalf("expected full time expression to stay intact, got %q", sentences[0])
	}
}

// BenchmarkExtractSmartSentences 基准测试句子切分性能
func BenchmarkExtractSmartSentences(b *testing.B) {
	text := "你好，这是一个用于测试句子切分性能的文本。它包含多个句子。每个句子都有不同的长度。我们希望测试在不同场景下的性能表现。"
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = ExtractSmartSentences(text, 2, 100, false)
	}
}

// BenchmarkExtractSmartSentences_LongText 测试长文本的切分性能
func BenchmarkExtractSmartSentences_LongText(b *testing.B) {
	longText := "这是一个很长的文本。"
	for i := 0; i < 10; i++ {
		longText += "它包含很多很多的句子。每个句子都有它自己的内容和意义。句子切分需要正确处理这些内容。"
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = ExtractSmartSentences(longText, 2, 200, false)
	}
}

func TestContainsSentenceSeparatorIgnoresStreamingTimeColon(t *testing.T) {
	if ContainsSentenceSeparator("现在是2026年3月20日 星期五 02:", false) {
		t.Fatal("expected trailing time colon not to trigger sentence split")
	}

	if !ContainsSentenceSeparator("现在是2026年3月20日 星期五 02:37:04。", false) {
		t.Fatal("expected final period to trigger sentence split")
	}
}
