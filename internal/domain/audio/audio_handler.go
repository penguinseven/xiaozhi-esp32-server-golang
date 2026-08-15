package audio

import (
	"xiaozhi-esp32-server-golang/internal/domain/audio/codec"
)

// AudioProcesser 音频处理器（适配 codec.FrameConverter）
type AudioProcesser struct {
	converter codec.FrameConverter
}

// GetAudioProcesser 创建音频处理器
func GetAudioProcesser(sampleRate int, channels int, perFrameDuration int) (*AudioProcesser, error) {
	conv, err := codec.New(codec.Config{
		SampleRate:       sampleRate,
		Channels:         channels,
		PerFrameDuration: perFrameDuration,
	})
	if err != nil {
		return nil, err
	}
	return &AudioProcesser{converter: conv}, nil
}

func (a *AudioProcesser) Decoder(audio []byte, pcmData []int16) (int, error) {
	return a.converter.Decode(audio, pcmData)
}

func (a *AudioProcesser) DecoderFloat32(audio []byte, pcmData []float32) (int, error) {
	return a.converter.DecodeFloat32(audio, pcmData)
}

func (a *AudioProcesser) Encoder(pcmData []int16, audio []byte) (int, error) {
	return a.converter.Encode(pcmData, audio)
}
