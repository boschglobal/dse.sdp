package file

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
)

type Handler interface {
	Detect(file string) any
	Import(ctx context.Context, file string, data any) error
}

var fileHandlers = map[string][]Handler{
	"text/yaml": {
		&kind.YamlKindHandler{},
	},
}

// GetHandler detects the type of file and returns the handler which supports
// it, together with the data detected by that handler.
func GetHandler(file string) (Handler, any, error) {
	mimeType, err := mimetype.DetectFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("unknown file type: %w", err)
	}
	fileType := strings.Split(mimeType.String(), ";")[0]
	switch strings.ToLower(filepath.Ext(file)) {
	case ".yaml", ".yml":
		fileType = "text/yaml"
	}
	slog.Debug("Detected file type", "file", file, "type", fileType)

	for _, h := range fileHandlers[fileType] {
		if data := h.Detect(file); data != nil {
			return h, data, nil
		}
	}
	return nil, nil, fmt.Errorf("unsupported file")
}
