package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GraphTestSuite struct {
	suite.Suite
	ctx     context.Context
	driver  neo4j.DriverWithContext
	session neo4j.SessionWithContext
}

func (s *GraphTestSuite) SetupSuite() {
	s.T().Log("Setup driver ...")
	s.ctx = context.Background()
	s.driver, _ = neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.BasicAuth("", "", ""),
		func(config *config.Config) {
			config.SocketConnectTimeout = 1 * time.Second
			config.MaxTransactionRetryTime = 1 * time.Second
		},
	)
}

func (s *GraphTestSuite) TearDownSuite() {
	s.driver.Close(s.ctx)
}

func (s *GraphTestSuite) SetupTest() {
	s.T().Log("Create session ...")
	s.session = s.driver.NewSession(s.ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
		// BoltLogger: neo4j.ConsoleBoltLogger(),
	})
	_, err := s.session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(s.ctx, "MATCH (n) DETACH DELETE n", map[string]any{})
	})
	require.NoError(s.T(), err)
}

func (s *GraphTestSuite) TearDownTest() {
	s.session.Close(s.ctx)
}

func TestGraphSuite(t *testing.T) {
	suite.Run(t, new(GraphTestSuite))
}

func (s *GraphTestSuite) TestGraphDrop() {
	t := s.T()

	// Add a node to the graph.
	s.session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(s.ctx, "CREATE (foo {name: 'Foo'})-[r:FOLLOWS]->(bar {name: 'Bar'})", map[string]any{})
	})

	// Run the command.
	cmd := NewGraphDropCommand("drop")
	args := []string{
		"--all",
	}
	err := cmd.Parse(args)
	assert.Nil(t, err)
	err = cmd.Run()
	assert.Nil(t, err)

	// Check the node count is 0.
	assert.Equal(t, int64(0), countGraphNodes(s))
}

func (s *GraphTestSuite) TestExport() {
	t := s.T()
	exportFile := filepath.Join(t.TempDir(), "export.cyp")

	// Add a node to the graph.
	s.session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(s.ctx, "CREATE (foo {name: 'Foo'})-[r:FOLLOWS]->(bar {name: 'Bar'})", map[string]any{})
	})

	// Run the command.
	cmd := NewGraphExportCommand("export")
	args := []string{
		exportFile,
	}
	err := cmd.Parse(args)
	assert.Nil(t, err)
	err = cmd.Run()
	assert.Nil(t, err)

	// Check the export was created.
	assert.FileExists(t, exportFile)
	exportContent, _ := os.ReadFile(exportFile)
	assert.Contains(t, string(exportContent), "CREATE (n:_IMPORT_ID {name: 'Foo',")
	assert.Contains(t, string(exportContent), "CREATE (n:_IMPORT_ID {name: 'Bar',")
	assert.Contains(t, string(exportContent), "CREATE (n)-[:FOLLOWS {}]->(m)")
}

func countGraphNodes(s *GraphTestSuite) int64 {
	t := s.T()
	count, _ := s.session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(s.ctx, "MATCH (n) RETURN count(*) AS count", map[string]any{})
		assert.NoError(t, err)
		record, err := result.Single(s.ctx)
		assert.NoError(t, err)
		count, _, err := neo4j.GetRecordValue[int64](record, "count")
		assert.NoError(t, err)
		return count, err
	})
	return count.(int64)
}
