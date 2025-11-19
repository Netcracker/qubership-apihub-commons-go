package scanner

import (
	"strings"

	"github.com/qubership-apihub-commons-go/api-spec-exposer/config"
)

// RestIdentifier identifies OpenAPI specifications
type RestIdentifier struct{}

func (i *RestIdentifier) CanHandle(path string) bool {
	ext := getFileExtension(path)
	return ext == "json" || ext == "yaml" || ext == "yml"
}

func (i *RestIdentifier) Identify(path string, content []byte) (*config.SpecMetadata, error) {
	var data map[string]interface{}
	var format config.Format
	var err error

	ext := getFileExtension(path)
	if ext == "json" {
		data, err = parseJSON(content)
		format = config.FormatJSON
	} else if ext == "yaml" || ext == "yml" {
		data, err = parseYAML(content)
		format = config.FormatYAML
	} else {
		return nil, nil
	}

	if err != nil {
		return nil, nil
	}

	openapiVersion := getString(data, "openapi")
	swaggerVersion := getString(data, "swagger")

	if openapiVersion == "" && swaggerVersion == "" {
		return nil, nil
	}

	if !hasKey(data, "info") {
		return nil, nil
	}

	name := getFileName(path)
	info, ok := data["info"].(map[string]interface{})
	if !ok || !hasKey(info, "title") {
		return nil, nil
	} else {
		if title, ok := info["title"].(string); ok && title != "" {
			name = title
		}
	}

	var docType config.DocumentType
	if strings.HasPrefix(openapiVersion, "3.1") {
		docType = config.DocTypeOpenAPI31
	} else if strings.HasPrefix(openapiVersion, "3.0") {
		docType = config.DocTypeOpenAPI30
	} else if strings.HasPrefix(swaggerVersion, "2.") {
		docType = config.DocTypeOpenAPI20
	} else {
		return nil, nil
	}

	//TODO: add x-api-kind value validation
	xApiKind := getString(data, "x-api-kind")
	if xApiKind == "" {
		xApiKind = getXApiKind(path)
	}

	return &config.SpecMetadata{
		Name:     name,
		FilePath: path,
		Type:     docType,
		ApiType:  config.ApiTypeRest,
		Format:   format,
		FileId:   generateFileId(path),
		XApiKind: xApiKind,
	}, nil
}
