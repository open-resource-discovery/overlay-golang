package marshaller

import (
	"strings"

	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"gopkg.in/yaml.v3"
)

func Marshal(media string, content any) (string, error) {
	if strings.HasPrefix(media, "text/xml") || strings.HasPrefix(media, "application/xml") {
		return utils.SafeCast[xml2json.Document](content).ToXML(), nil
	}

	if strings.HasPrefix(media, "text/yaml") || strings.HasPrefix(media, "application/yaml") {
		result, err := yaml.Marshal(content)

		return string(utils.Ternary(result == nil, []byte{}, result)), err
	}

	if strings.HasPrefix(media, "application/json") || strings.HasPrefix(media, "text/json") {
		return oj.JSON(content), nil
	}

	return "", errors.Errorf("unsupported media type: %s", media)
}

func Unmarshal(media string, content string) (any, error) {
	if strings.HasPrefix(media, "text/xml") || strings.HasPrefix(media, "application/xml") {
		return xml2json.Convert(content)
	}

	if strings.HasPrefix(media, "text/yaml") || strings.HasPrefix(media, "application/yaml") {
		result := make(map[string]any)

		return result, yaml.Unmarshal([]byte(content), result)
	}

	if strings.HasPrefix(media, "application/json") || strings.HasPrefix(media, "text/json") {
		return oj.ParseString(content)
	}

	return nil, errors.Errorf("unsupported media type: %s", media)
}
