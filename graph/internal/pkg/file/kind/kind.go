package kind

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"path/filepath"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"gopkg.in/yaml.v3"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type YamlKindHandler struct{}

func (h *YamlKindHandler) Detect(file string) any {
	decoder, err := func(file string) (*yaml.Decoder, error) {
		yamlFile, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer yamlFile.Close()

		data, _ := io.ReadAll(yamlFile)
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		return decoder, nil
	}(file)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	docList := []KindDoc{}
	for {
		var doc KindDoc
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		doc.file = filepath.Base(file)
		docList = append(docList, doc)
		fmt.Printf("  Handler:  yaml/kind=%s\n", doc.Kind)
	}

	if len(docList) != 0 {
		return docList
	}
	return nil
}

func (h *YamlKindHandler) Import(ctx context.Context, file string, data any) {
	if data == nil {
		fmt.Println("Error: no data object to import!")
		return
	}
	docList := data.([]KindDoc)

	session, _ := graph.Session(ctx)
	defer session.Close(ctx)

	docIndex := 1
	for _, kd := range docList {
		fmt.Println(kd.file)
		fmt.Println(kd.Kind)
		simulationSpec, _ := kd.Spec.(*SimulationSpec)
		// FILE -[:CONTAINS]-> KINDDOC
		kind_props := map[string]any{
			"labels":      kd.Metadata.Labels,
			"annotations": kd.Metadata.Annotations,
		}
		var ast_props map[string]any
		if kd.Kind == "Simulation" {
			ast_props = map[string]any{
				"labels":      kd.Metadata.Labels,
				"annotations": kd.Metadata.Annotations,
				"arch":      simulationSpec.Arch,
			}
		}
		file_id, _ := graph.Node(ctx, session, []string{"File"}, kd.file)
		if kd.Kind == "SignalGroup" {
			var b strings.Builder
			properties := map[string]any{
				"signalgroup_name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:SignalGroup {signalgroup_name: $signalgroup_name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Stack" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Stack {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		}  else if kd.Kind == "Model" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Model {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Network" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Network {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Runnable" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Runnable {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Manifest" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Manifest {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Propagator" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:Propagator {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "Simulation" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    ast_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Ast:Simulation {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else if kd.Kind == "ParameterSet" {
			var b strings.Builder
			properties := map[string]any{
				"name":     kd.Metadata.Name,
				"filename": kd.file,
				"index":    strconv.FormatInt(int64(docIndex), 10),
				"props":    kind_props,
			}
			b.WriteString("MATCH (f:File {name: $filename}) ")
			b.WriteString("MERGE (f)-[r:Contains {index: $index}]->(n:Sim:ParameterSet {name: $name}) ")
			b.WriteString("ON CREATE SET n += $props ")
			b.WriteString("ON MATCH SET n += $props ")
			b.WriteString("RETURN id(n) AS id")
			kind_id, _ := graph.Query(ctx, session, b.String(), properties)
			kd.kind_id = kind_id
		} else {
			match_props := map[string]string{
				"name": kd.Metadata.Name,
			}
			kind_id, _ := graph.NodeExt(ctx, session, []string{"KindDoc", kd.Kind}, match_props, kind_props)
			graph.Relation(ctx, session, file_id, kind_id, []string{"Contains"})
			kd.kind_id = kind_id
		}

		// KINDDOC -->
		kd.CreateGraph(ctx, session)

		// Next doc.
		docIndex += 1
	}
}

type KindSpec interface {
	MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error
}

type KindDoc struct {
	kind_id  int64
	file     string
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Annotations map[string]interface{} `yaml:"annotations"`
		Labels      map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec interface{} `yaml:"-"`
}

func SetSpec(kd *KindDoc) error {
	switch kd.Kind {
	// case "Manifest":
	// 	kd.Spec = newManifestSpec()
	case "SignalGroup":
		kd.Spec = newSignalGroupSpec()
	case "Model":
		kd.Spec = newModelSpec()
	case "Stack":
		kd.Spec = newStackSpec()
	case "Network":
		kd.Spec = newNetworkSpec()
	case "Runnable":
		kd.Spec = newRunnableSpec()
	case "Manifest":
		kd.Spec = newManifestSpec()
	case "Propagator":
		kd.Spec = newPropagatorspec()
	case "Simulation":
		kd.Spec = newSimulationSpec()
	case "ParameterSet":
		kd.Spec = newParameterSetSpec()
	default:
		return fmt.Errorf("unknown kind: %s", kd.Kind)
	}
	return nil
}

func (kd KindDoc) CreateGraph(ctx context.Context, session neo4j.SessionWithContext) {
	ks := kd.Spec.(KindSpec)
	ks.MergeGraph(ctx, session, &kd)
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
