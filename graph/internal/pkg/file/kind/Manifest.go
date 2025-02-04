package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

type ManifestSpec kind.ManifestSpec

func newManifestSpec() *ManifestSpec {
	return new(ManifestSpec)
}

func (m *ManifestSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	manifest_id := kd.kind_id

	// MANIFEST -[HAS]-> DOCUMENTATION
	if m.Documentation != nil {
		for _, doc := range *m.Documentation {
			matchProps := map[string]string{"name": doc.Name}
			nodeProps := map[string]any{
				"generate":   doc.Generate,
				"modelc":     doc.Modelc,
				"processing": doc.Processing,
				"repo":       doc.Repo,
				"uri":        doc.Uri,
			}
			doc_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Documentation"}, matchProps, nodeProps)
			graph.Relation(ctx, session, manifest_id, doc_id, []string{"Has"})
		}
	}

	// MANIFEST -[HAS]-> MODEL
	for _, model := range m.Models {
		matchProps := map[string]string{"name": model.Name}
		nodeProps := map[string]any{
			"arch":     model.Arch,
			"repo":     model.Repo,
			"schema":   model.Schema,
			"version":  model.Version,
		}
		model_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Model"}, matchProps, nodeProps)
		graph.Relation(ctx, session, manifest_id, model_id, []string{"Has"})

		if model.Channels != nil {
			for _, c := range *model.Channels {
				alias := ""
				if c.Alias != nil {
					alias = *c.Alias
				}

				match_props := map[string]string{
					"alias": alias,
				}
				node_props := map[string]any{
					"selectors": c.Selectors,
				}
				channel_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Channels"}, match_props, node_props)
				graph.Relation(ctx, session, model_id, channel_id, []string{"Has"})
			}
		}
	}

	// MANIFEST -[HAS]-> REPO
	for _, repo := range m.Repos {
		matchProps := map[string]string{"name": repo.Name}
		nodeProps := map[string]any{
			"path":     repo.Path,
			"registry": repo.Registry,
			"repo":     repo.Repo,
			"token":    repo.Token,
			"user":     repo.User,
		}
		repo_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Repo"}, matchProps, nodeProps)
		graph.Relation(ctx, session, manifest_id, repo_id, []string{"Has"})
	}

	// MANIFEST -[HAS]-> SIMULATION
	for _, sim := range m.Simulations {
		matchProps := map[string]string{"name": sim.Name}
		var sim_id int64

		if sim.Parameters != nil {
			nodeProps := map[string]any{
				"transport":   sim.Parameters.Transport,
				"environment": sim.Parameters.Environment,
			}
			sim_id, _ = graph.NodeExt(ctx, session, []string{"Sim:Simulation"}, matchProps, nodeProps)
			graph.Relation(ctx, session, manifest_id, sim_id, []string{"Has"})
		}

		if sim.Files != nil {
			// SIMULATION -[HAS]-> FILES
			for _, file := range *sim.Files {
				fileMatchProps := map[string]string{"name": file.Name}
				fileNodeProps := map[string]any{
					"generate":   file.Generate,
					"modelc":     file.Modelc,
					"processing": file.Processing,
					"repo":       file.Repo,
					"uri":        file.Uri,
				}
				file_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Files"}, fileMatchProps, fileNodeProps)
				graph.Relation(ctx, session, sim_id, file_id, []string{"Has"})
			}
		}

		if sim.Models != nil {
			// SIMULATION -[HAS]-> MODELS
			for _, model := range sim.Models {
				modelMatchProps := map[string]string{"name": model.Name}
				modelNodeProps := map[string]any{
					"model":    model.Model,
				}
				model_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Models"}, modelMatchProps, modelNodeProps)
				graph.Relation(ctx, session, sim_id, model_id, []string{"Has"})

				if model.Channels != nil {
					for _, c := range model.Channels {
						alias := ""
						if c.Alias != nil {
							alias = *c.Alias
						}

						match_props := map[string]string{
							"alias": alias,
						}
						node_props := map[string]any{
							"selectors": c.Selectors,
						}
						channel_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Channels"}, match_props, node_props)
						graph.Relation(ctx, session, model_id, channel_id, []string{"Has"})
					}
				}
			}
		}
	}

	// MANIFEST -[HAS]-> TOOL
	for _, tool := range m.Tools {
		matchProps := map[string]string{"name": tool.Name}
		nodeProps := map[string]any{
			"arch":    tool.Arch,
			"repo":    tool.Repo,
			"schema":  tool.Schema,
			"version": tool.Version,
		}
		tool_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Tool"}, matchProps, nodeProps)
		graph.Relation(ctx, session, manifest_id, tool_id, []string{"Has"})
	}

	return nil
}
