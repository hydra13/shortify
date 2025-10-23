package shortidgenerator

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hydra13/shortify/internal/config"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g Generator) GenerateShortID(url string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v-%v", url, time.Now())))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return encoded[:config.KeyLength]
}
