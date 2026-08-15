package codec

import "errors"

// FrameConverter 音频帧编解码接口（领域层抽象，隔离 CGo 依赖）
type FrameConverter interface {
	// Decode 解码 opus 帧为 PCM int16
	Decode(opus []byte, pcm []int16) (int, error)
	// DecodeFloat32 解码 opus 帧为 PCM float32
	DecodeFloat32(opus []byte, pcm []float32) (int, error)
	// Encode 编码 PCM int16 为 opus 帧
	Encode(pcm []int16, opus []byte) (int, error)
	// Close 释放资源
	Close() error
}

// Config 编解码器配置
type Config struct {
	SampleRate       int
	Channels         int
	PerFrameDuration int
}

// New 创建 FrameConverter（自动选择 cgo/纯 Go 实现）
func New(cfg Config) (FrameConverter, error) {
	if cfg.SampleRate <= 0 || cfg.Channels <= 0 {
		return nil, errors.New("codec: invalid sample rate or channels")
	}
	return newConverter(cfg)
}
