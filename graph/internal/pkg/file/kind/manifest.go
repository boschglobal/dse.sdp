package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type ManifestSpec schema_kind.ManifestSpec

func newManifestSpec() *ManifestSpec {
	return new(ManifestSpec)
}

func (m *ManifestSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// MANIFEST -[HAS]-> DOCUMENTATION
	each(m.Documentation, func(doc schema_kind.File) {
		mergeManifestFile(w, id, "Sim:Documentation", doc)
	})

	// MANIFEST -[HAS]-> MODEL
	for _, model := range m.Models {
		modelID := w.has(id, "Sim:Model",
			map[string]string{"name": model.Name},
			map[string]any{
				"arch":    model.Arch,
				"repo":    model.Repo,
				"schema":  model.Schema,
				"version": model.Version,
			})
		each(model.Channels, func(c schema_kind.Channel) {
			mergeManifestChannel(w, modelID, c)
		})
	}

	// MANIFEST -[HAS]-> REPO
	for _, repo := range m.Repos {
		w.has(id, "Sim:Repo",
			map[string]string{"name": repo.Name},
			map[string]any{
				"path":     repo.Path,
				"registry": repo.Registry,
				"repo":     repo.Repo,
				"token":    repo.Token,
				"user":     repo.User,
			})
	}

	// MANIFEST -[HAS]-> SIMULATION
	for _, sim := range m.Simulations {
		mergeManifestSimulation(w, id, sim)
	}

	// MANIFEST -[HAS]-> TOOL
	for _, tool := range m.Tools {
		w.has(id, "Sim:Tool",
			map[string]string{"name": tool.Name},
			map[string]any{
				"arch":    tool.Arch,
				"repo":    tool.Repo,
				"schema":  tool.Schema,
				"version": tool.Version,
			})
	}
	return w.err
}

func mergeManifestSimulation(w *writer, manifestID int64, sim schema_kind.Simulation) {
	props := map[string]any{}
	if sim.Parameters != nil {
		props["transport"] = sim.Parameters.Transport
		props["environment"] = sim.Parameters.Environment
	}
	simID := w.has(manifestID, "Sim:Simulation", map[string]string{"name": sim.Name}, props)

	// SIMULATION -[HAS]-> FILES
	each(sim.Files, func(f schema_kind.File) {
		mergeManifestFile(w, simID, "Sim:Files", f)
	})

	// SIMULATION -[HAS]-> MODELS
	for _, model := range sim.Models {
		modelID := w.has(simID, "Sim:Models",
			map[string]string{"name": model.Name},
			map[string]any{"model": model.Model})
		for _, c := range model.Channels {
			mergeManifestChannel(w, modelID, c)
		}
	}
}

func mergeManifestFile(w *writer, parent int64, label string, f schema_kind.File) {
	w.has(parent, label,
		map[string]string{"name": f.Name},
		map[string]any{
			"generate":   f.Generate,
			"modelc":     f.Modelc,
			"processing": f.Processing,
			"repo":       f.Repo,
			"uri":        f.Uri,
		})
}

func mergeManifestChannel(w *writer, parent int64, c schema_kind.Channel) {
	w.has(parent, "Sim:Channels",
		map[string]string{"alias": deref(c.Alias)},
		map[string]any{"selectors": c.Selectors})
}
