package structs

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	validRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idxBits    = 6
	idxMask    = 1<<idxBits - 1
)

//see http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func SessionID(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & idxMask); idx < len(validRunes) {
			b[i] = validRunes[idx]
			i++
		}
	}
	return string(b)
}
