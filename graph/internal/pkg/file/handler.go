package file

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
	"github.com/gabriel-vasile/mimetype"
)

type Handler interface {
	Detect(file string) any
	Import(ctx context.Context, file string, data any)
}

func GetHandler(file string) (Handler, any, error) {

	// Add yaml detection ;-)
	//yamlDetector := func(raw []byte, limit uint32) bool {
	//	fmt.Println("yaml detector")
	//	return false
	//}
	//mimetype.Lookup("text/plain").Extend(yamlDetector, "text/yaml", ".yaml")

	// Detect the file type and handler.
	mimeType, err := mimetype.DetectFile(file)
	if err != nil {
		fmt.Println(err)
		return nil, nil, fmt.Errorf("unknown file type")
	}
	fileType := strings.Split(mimeType.String(), ";")[0]
	yamlExtensions := map[string]bool{
		".yaml": true,
		".yml":  true,
	}
	if yamlExtensions[strings.ToLower(filepath.Ext(file))] {
		fileType = "text/yaml"
	}
	fmt.Println("  Type: ", fileType)

	fileHandlers := map[string][]Handler{
		// "text/csv": {
		// 	&CsvInputModelHandler{},
		// },
		// "text/xml": {
		// 	&XmlFmuHandler{},
		// },
		// "text/plain": {
		// 	&JsonSoftecuM2eE2mHandler{},
		// },
		"text/yaml": {
			&kind.YamlKindHandler{},
		},
	}
	handler, data, err := func(file string, handlers []Handler) (Handler, any, error) {
		for _, h := range handlers {
			data := h.Detect(file)
			if data != nil {
				return h, data, nil
			}
		}
		return nil, nil, fmt.Errorf("unsupported file")
	}(file, fileHandlers[fileType])

	// Return the selected handler.
	if err != nil {
		fmt.Println("  Handler: ", "<unknown>")
		return nil, nil, err
	}
	return handler, data, nil
}
