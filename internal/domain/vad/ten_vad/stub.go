//go:build !cgo || !ten_vad

package ten_vad

import (
	"errors"

	. "xiaozhi-esp32-server-golang/internal/domain/vad/inter"
)

// TenVAD stub（CGo/ten_vad 不可用时）
type TenVAD struct{}

func (t *TenVAD) Close() error { return nil }
func (t *TenVAD) IsValid() bool { return false }
func (t *TenVAD) Reset() error { return nil }
func (t *TenVAD) IsVAD(pcmData []float32) (bool, error) {
	return false, errors.New("ten_vad: CGo disabled")
}
func (t *TenVAD) IsVADExt(pcmData []float32, sampleRate int, frameSize int) (bool, error) {
	return false, errors.New("ten_vad: CGo disabled")
}

// AcquireVAD stub
func AcquireVAD(config map[string]interface{}) (VAD, error) {
	return nil, errors.New("ten_vad: CGo disabled, use silero_vad instead")
}

// ReleaseVAD stub
func ReleaseVAD(vad VAD) error {
	return nil
}
