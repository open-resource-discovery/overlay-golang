//go:build unit || integration

package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

func LoadFixture(path string) string {
	bytes, err := os.ReadFile(filepath.Join(utils.First(os.Getwd()), path))
	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %s", path, err.Error()))
	}

	return string(bytes)
}

func UnmarshalFixture[T any](path string) T {
	result, err := marshaller.Unmarshal(fmt.Sprintf("application/%s", strings.Split(filepath.Base(path), ".")[1]), LoadFixture(path))
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s: %s", path, err.Error()))
	}

	return result.(T)
}
