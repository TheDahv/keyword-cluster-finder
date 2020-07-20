package graph

import (
	"fmt"

	"github.com/jamesneve/go-markov-cluster/graph"
	"github.com/thedahv/keyword-cluster-finder/pkg/rankings"
	"github.com/thedahv/keyword-cluster-finder/pkg/rbo"
)

// Graph builds a network of keywords and their relationship to other keywords
// with respect to the similarity of their SERPs
type Graph struct {
	rboPValue            float64
	clusterPower         int
	clusterInflation     int
	maxComputeIterations int
}

// Option configures a graph
type Option func(g *Graph)

// WithRBOPValue configures the graph to use the specified RBO probability value
func WithRBOPValue(p float64) Option {
	return func(g *Graph) {
		g.rboPValue = p
	}
}

// WithClusterPower configures the graph to use the specified power value to
// compute graph clusters
func WithClusterPower(p int) Option {
	return func(g *Graph) {
		g.clusterPower = p
	}
}

// WithClusterInflation configures the graph to use the specified inflation
// value to compute graph clusters
func WithClusterInflation(i int) Option {
	return func(g *Graph) {
		g.clusterInflation = i
	}
}

// WithClusterMaxIterations configures the graph to limit the number of
// computation steps used when computing graph clusters
func WithClusterMaxIterations(i int) Option {
	return func(g *Graph) {
		g.maxComputeIterations = i
	}
}

// New creates a new Graph configured by options
func New(options ...Option) *Graph {
	g := &Graph{
		rboPValue:            0.9,
		clusterPower:         2,
		clusterInflation:     5,
		maxComputeIterations: 100,
	}

	for _, o := range options {
		o(g)
	}

	return g
}

// ClusterGroup is a cluster of highly-related keywords with respect to the
// similarity of their SERP members
type ClusterGroup struct {
	Name     string
	Keywords []string
}

// FindClusters adds gathered SERP data to a graph, computes the RBO weights
// among all SERPs, and returns clusters of keywords whose SERP members are
// similar
func (g Graph) FindClusters(kd rankings.KeywordData) ([]ClusterGroup, error) {
	_g := graph.NewGraph()
	nodes := make(map[string]*graph.Node)
	for keyword := range kd {
		n := graph.NewNode(keyword)
		_g.AddNode(&n)
		nodes[keyword] = &n
	}

	for fromKeyword, fromSERP := range kd {
		for toKeyword, toSERP := range kd {
			if fromKeyword == toKeyword {
				continue
			}

			_, _, rboExt, err := rbo.RBO(fromSERP, toSERP, g.rboPValue)
			if err != nil {
				return nil, fmt.Errorf("error computing %s->%s: %v", fromKeyword, toKeyword, err)
			}
			_g.AddEdge(nodes[fromKeyword], nodes[toKeyword], rboExt)
		}
	}

	c, err := _g.GetClusters(g.clusterPower, g.clusterInflation, g.maxComputeIterations)
	if err != nil {
		return nil, fmt.Errorf("could not find graph clusters: %v", err)
	}

	var clusters []ClusterGroup
	for _, cluster := range *c {
		name := getShortestKeyword(cluster)
		clusters = append(clusters, ClusterGroup{
			Name:     name,
			Keywords: cluster,
		})
	}

	return clusters, nil
}

func getShortestKeyword(keywords []string) string {
	shortest := keywords[0]
	for i := 1; i < len(keywords); i++ {
		if len(keywords[i]) < len(shortest) {
			shortest = keywords[i]
		}
	}

	return shortest
}
