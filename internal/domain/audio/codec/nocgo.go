//go:build !cgo

package codec

import "errors"

type passthroughConverter struct{}

func newConverter(cfg Config) (FrameConverter, error) {
	return &passthroughConverter{}, nil
}

func (c *passthroughConverter) Decode(opusData []byte, pcm []int16) (int, error) {
	return 0, errors.New("codec: opus decode not available (CGo disabled)")
}

func (c *passthroughConverter) DecodeFloat32(opusData []byte, pcm []float32) (int, error) {
	return 0, errors.New("codec: opus decode not available (CGo disabled)")
}

func (c *passthroughConverter) Encode(pcm []int16, opusData []byte) (int, error) {
	return 0, errors.New("codec: opus encode not available (CGo disabled)")
}

func (c *passthroughConverter) Close() error {
	return nil
}
