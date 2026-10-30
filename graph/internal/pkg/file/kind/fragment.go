package kind

import "errors"

// mergeAnnotationGraph creates the nodes described by the "graph" annotation of
// a kind document, each linked to the document node id by the declared edge.
func mergeAnnotationGraph(w *writer, kd *KindDoc, id int64) {
	fragment, _ := kd.Metadata.Annotations["graph"].(map[string]any)
	nodes, _ := fragment["nodes"].([]any)
	if len(nodes) == 0 {
		return
	}

	edge, _ := fragment["edge"].(map[string]any)
	label, _ := edge["label"].(string)
	direction, _ := edge["direction"].(string)
	if label == "" {
		w.check(errors.New("graph annotation: edge label missing"))
		return
	}

	for _, n := range nodes {
		node, ok := n.(map[string]any)
		if !ok {
			continue
		}
		nodeLabel, _ := node["label"].(string)
		if nodeLabel == "" {
			continue
		}
		match := map[string]string{}
		props := map[string]any{}
		properties, _ := node["properties"].(map[string]any)
		for k, v := range properties {
			if s, ok := v.(string); ok {
				match[k] = s
			} else {
				props[k] = v
			}
		}

		nodeID := w.node("Sim:"+nodeLabel, match, props)
		if direction == "in" {
			w.relate(nodeID, id, label, nil)
		} else {
			w.relate(id, nodeID, label, nil)
		}
	}
}
