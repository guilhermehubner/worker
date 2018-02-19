package worker

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/guilhermehubner/worker/log"
)

func makeIdentifier() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Get().Error(err)
		return ""
	}
	return fmt.Sprintf("%x", b)
}
