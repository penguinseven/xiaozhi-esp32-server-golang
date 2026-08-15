//go:build cgo

package codec

import (
	"gopkg.in/hraban/opus.v2"
)

type opusConverter struct {
	decoder *opus.Decoder
	encoder *opus.Encoder
}

func newConverter(cfg Config) (FrameConverter, error) {
	decoder, err := opus.NewDecoder(cfg.SampleRate, cfg.Channels)
	if err != nil {
		return nil, err
	}
	encoder, err := opus.NewEncoder(cfg.SampleRate, cfg.Channels, opus.AppAudio)
	if err != nil {
		return nil, err
	}
	return &opusConverter{
		decoder: decoder,
		encoder: encoder,
	}, nil
}

func (c *opusConverter) Decode(opusData []byte, pcm []int16) (int, error) {
	return c.decoder.Decode(opusData, pcm)
}

func (c *opusConverter) DecodeFloat32(opusData []byte, pcm []float32) (int, error) {
	return c.decoder.DecodeFloat32(opusData, pcm)
}

func (c *opusConverter) Encode(pcm []int16, opusData []byte) (int, error) {
	return c.encoder.Encode(pcm, opusData)
}

func (c *opusConverter) Close() error {
	return nil
}
