package util

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateUuidSample() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
