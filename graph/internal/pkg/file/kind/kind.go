package kind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

var errUnknownKind = errors.New("unknown kind")

// kindDef describes how a kind is decoded and stored in the graph.
type kindDef struct {
	labels  string // Node labels of the kind document.
	nameKey string // Property used to match the node by name.
	newSpec func() KindSpec
}

var kindDefs = map[string]kindDef{
	"SignalGroup":  {"Sim:SignalGroup", "signalgroup_name", func() KindSpec { return newSignalGroupSpec() }},
	"Stack":        {"Sim:Stack", "name", func() KindSpec { return newStackSpec() }},
	"Model":        {"Sim:Model", "name", func() KindSpec { return newModelSpec() }},
	"Network":      {"Sim:Network", "name", func() KindSpec { return newNetworkSpec() }},
	"Runnable":     {"Sim:Runnable", "name", func() KindSpec { return newRunnableSpec() }},
	"Manifest":     {"Sim:Manifest", "name", func() KindSpec { return newManifestSpec() }},
	"Propagator":   {"Sim:Propagator", "name", func() KindSpec { return newPropagatorSpec() }},
	"Simulation":   {"Ast:Simulation", "name", func() KindSpec { return newSimulationSpec() }},
	"ParameterSet": {"Sim:ParameterSet", "name", func() KindSpec { return newParameterSetSpec() }},
}

// YamlKindHandler detects and imports YAML files made of kind documents.
type YamlKindHandler struct {
	// Out receives progress output, defaults to os.Stdout.
	Out io.Writer
}

func (h *YamlKindHandler) out() io.Writer {
	if h.Out == nil {
		return os.Stdout
	}
	return h.Out
}

func (h *YamlKindHandler) Detect(file string) any {
	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Read file", "file", file, "err", err)
		return nil
	}

	var docList []KindDoc
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var doc KindDoc
		err := decoder.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			var typeErr *yaml.TypeError
			if errors.Is(err, errUnknownKind) || errors.As(err, &typeErr) {
				continue // Skip this document.
			}
			slog.Error("Decode file", "file", file, "err", err)
			break // Decoder cannot make further progress.
		}
		doc.file = filepath.Base(file)
		docList = append(docList, doc)
		_, _ = fmt.Fprintf(h.out(), "  Handler:  yaml/kind=%s\n", doc.Kind)
	}

	if len(docList) == 0 {
		return nil
	}
	return docList
}

// Import merges the detected documents into the graph, one transaction per
// document. All documents are attempted, the errors of failed ones are joined.
func (h *YamlKindHandler) Import(ctx context.Context, file string, data any) error {
	docList, ok := data.([]KindDoc)
	if !ok || len(docList) == 0 {
		return fmt.Errorf("no data object to import from %q", file)
	}

	session, err := graph.Session(ctx)
	if err != nil {
		return err
	}
	defer session.Close(ctx)

	var errs []error
	for i, kd := range docList {
		slog.Info("Import kind", "file", kd.file, "kind", kd.Kind)
		err := graph.WriteTx(ctx, session, func(ex graph.Executor) error {
			return kd.merge(ctx, ex, i+1)
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %s document %d: %w", kd.file, kd.Kind, i+1, err))
		}
	}
	return errors.Join(errs...)
}

// merge creates the graph nodes of a kind document, with docIndex its position in the file.
func (kd *KindDoc) merge(ctx context.Context, ex graph.Executor, docIndex int) error {
	id, err := kd.mergeNode(ctx, ex, docIndex)
	if err != nil {
		return err
	}
	kd.kindID = id
	return kd.Spec.(KindSpec).MergeGraph(ctx, ex, kd)
}

// mergeNode creates the node of a kind document, linked from its file node,
// and returns the node id.
func (kd *KindDoc) mergeNode(ctx context.Context, ex graph.Executor, docIndex int) (int64, error) {
	def, ok := kindDefs[kd.Kind]
	if !ok {
		return -1, fmt.Errorf("%w: %s", errUnknownKind, kd.Kind)
	}
	if _, err := graph.Node(ctx, ex, []string{"File"}, kd.file); err != nil {
		return -1, err
	}

	props := map[string]any{
		"labels":      kd.Metadata.Labels,
		"annotations": kd.Metadata.Annotations,
	}
	if spec, ok := kd.Spec.(*SimulationSpec); ok {
		props["arch"] = spec.Arch
	}
	query := "MATCH (f:File {name: $filename}) " +
		"MERGE (f)-[r:Contains {index: $index}]->(n:" + def.labels + " {" + def.nameKey + ": $name}) " +
		"ON CREATE SET n += $props " +
		"ON MATCH SET n += $props " +
		"RETURN id(n) AS id"
	id, err := graph.Query(ctx, ex, query, map[string]any{
		"name":     kd.Metadata.Name,
		"filename": kd.file,
		"index":    strconv.Itoa(docIndex),
		"props":    props,
	})
	if err != nil {
		return -1, err
	}
	if id < 0 {
		return -1, fmt.Errorf("kind node for %q not created", kd.Metadata.Name)
	}
	return id, nil
}

type KindSpec interface {
	MergeGraph(ctx context.Context, ex graph.Executor, kd *KindDoc) error
}

type KindDoc struct {
	kindID   int64
	file     string
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string                 `yaml:"name"`
		Annotations map[string]interface{} `yaml:"annotations"`
		Labels      map[string]string      `yaml:"labels"`
	} `yaml:"metadata"`
	Spec interface{} `yaml:"-"`
}

// SetSpec allocates the spec matching the kind of the document.
func SetSpec(kd *KindDoc) error {
	def, ok := kindDefs[kd.Kind]
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownKind, kd.Kind)
	}
	kd.Spec = def.newSpec()
	return nil
}

func (kd *KindDoc) UnmarshalYAML(n *yaml.Node) error {
	type KD KindDoc
	type T struct {
		*KD  `yaml:",inline"`
		Spec yaml.Node `yaml:"spec"`
	}
	// Create the combined struct (i.e. outer KindDoc + Spec) and then decode
	// the yaml node. Kd will be populated, and then T.Spec can be specifically
	// decoded into the kd.Spec based on the kind.
	obj := &T{KD: (*KD)(kd)}
	if err := n.Decode(obj); err != nil {
		return err
	}
	if err := SetSpec(kd); err != nil {
		return err
	}
	return obj.Spec.Decode(kd.Spec)
}
