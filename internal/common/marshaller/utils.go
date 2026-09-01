package marshaller

import (
	"strings"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"gopkg.in/yaml.v3"
)

func MustMarshal(media string, content any) string {
	result, err := Marshal(media, content)

	if err != nil {
		panic(errors.WrapPrefix(errors.Wrap(err, errors.Severity_Fatal), "failed to marshal content as %s", media))
	}

	return result
}

func MustUnmarshal(media string, content string) any {
	result, err := Unmarshal(media, content)

	if err != nil {
		panic(errors.WrapPrefix(errors.Wrap(err, errors.Severity_Fatal), "failed to unmarshal content as %s", media))
	}

	return result
}

func Marshal(media string, content any) (string, *errors.OverlayError) {
	if strings.HasPrefix(media, "text/xml") || strings.HasPrefix(media, "application/xml") {
		return utils.SafeCast[xml2json.Document](content).ToXML(), nil
	}

	if strings.HasPrefix(media, "text/yaml") || strings.HasPrefix(media, "application/yaml") {
		result, err := yaml.Marshal(content)

		return string(utils.Ternary(result == nil, []byte{}, result)), errors.Wrap(err, errors.Severity_Error)
	}

	if strings.HasPrefix(media, "application/json") || strings.HasPrefix(media, "text/json") {
		return oj.JSON(content), nil
	}

	return "", errors.Create(errors.Severity_Error, "unsupported media type: %s", media)
}

func Unmarshal(media string, content string) (any, *errors.OverlayError) {
	if strings.HasPrefix(media, "text/xml") || strings.HasPrefix(media, "application/xml") {
		return xml2json.Convert(content)
	}

	if strings.HasPrefix(media, "text/yaml") || strings.HasPrefix(media, "application/yaml") {
		result := make(map[string]any)

		return result, errors.Wrap(yaml.Unmarshal([]byte(content), result), errors.Severity_Error)
	}

	if strings.HasPrefix(media, "application/json") || strings.HasPrefix(media, "text/json") {
		result, err := oj.ParseString(content)

		return result, errors.Wrap(err, errors.Severity_Error)
	}

	return "", errors.Create(errors.Severity_Error, "unsupported media type: %s", media)
}
