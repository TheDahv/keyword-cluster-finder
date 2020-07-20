package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thedahv/keyword-cluster-finder/pkg/graph"
	"github.com/thedahv/keyword-cluster-finder/pkg/rankings"
)

// WeightedEdge describes a relationship between 2 SERPs where its weight
// reflects its similarity with respect to the RBO calculation
type WeightedEdge struct {
	From rankings.SERP
	To   rankings.SERP
	RBO  float64
}

const rboPValue = 0.9

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatal("rankings directory argument required")
	}

	directory := args[0]
	kd, err := rankings.ProcessDirectory(directory)
	if err != nil {
		log.Fatalf("could not process directory: %v", err)
	}

	g := graph.New(
		graph.WithRBOPValue(0.9),
		graph.WithClusterPower(2),
		graph.WithClusterInflation(5),
		graph.WithClusterMaxIterations(100),
	)
	clusters, err := g.FindClusters(kd)
	if err != nil {
		log.Fatalf("could not find graph clusters: %v", err)
	}

	for _, cluster := range clusters {
		fmt.Printf("Cluster: '%s'\n", cluster.Name)
		for _, kw := range cluster.Keywords {
			fmt.Printf("\t%s\n", kw)
		}
	}
}
