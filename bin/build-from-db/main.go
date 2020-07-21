package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cheggaaa/pb"
	"github.com/thedahv/keyword-cluster-finder/pkg/data"
	"github.com/thedahv/keyword-cluster-finder/pkg/graph"
	"github.com/thedahv/keyword-cluster-finder/pkg/rankings"
)

const rboPValue = 0.9

func main() {
	// Required
	var domainID = flag.Int("domainID", 0, "Domain ID")
	var configPath = flag.String("config", "", "app JSON config")

	// Optional
	var p = flag.Float64("p", 0.9, "RBO p value")
	var pow = flag.Int("pow", 5, "Cluster power")
	var inf = flag.Int("inf", 2, "Cluster inflation")
	flag.Parse()

	if *domainID == 0 {
		log.Fatalf("must provide a domain ID")
	}
	if *configPath == "" {
		log.Fatalf("must provide config path")
	}

	fmt.Println("parsing config...")
	conf, err := parseConfig(*configPath)
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	fmt.Println()
	fmt.Println("connecting to database...")
	driver, err := data.New(
		data.WithUserAndPass(conf.DB.User, conf.DB.Pass),
		data.WithHost(conf.DB.Host),
		data.WithDatabase(conf.DB.Database),
	)
	if err != nil {
		log.Fatalf("could not set up database connection: %v", err)
	}

	fmt.Println()
	fmt.Println("fetching keywords...")
	keywords, err := driver.FetchKeywords(*domainID)
	if err != nil {
		log.Fatalf("could not read keywords: %v", err)
	}
	fmt.Printf("got %d keywords\n\n", len(keywords))

	fmt.Println("querying database...")
	bar := pb.StartNew(len(keywords))

	kd := rankings.New()
	err = kd.BuildFromDatabase(driver, *domainID, keywords, bar)
	if err != nil {
		log.Fatalf("could not build from database: %v", err)
	}
	bar.Finish()

	g := graph.New(
		graph.WithRBOPValue(*p),
		graph.WithClusterPower(*pow),
		graph.WithClusterInflation(*inf),
		graph.WithClusterMaxIterations(100),
	)
	fmt.Println()
	fmt.Println("finding graph...")
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

func parseConfig(path string) (config, error) {
	var c config
	f, err := os.Open(path)
	if err != nil {
		return c, fmt.Errorf("could not open file: %v", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return c, fmt.Errorf("could not read file: %v", err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("could not parse config: %v", err)
	}

	return c, nil
}

func readKeywords(path string) ([]string, error) {
	var keywords []string
	f, err := os.Open(path)
	if err != nil {
		return keywords, fmt.Errorf("could not open file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		keywords = append(keywords, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return keywords, fmt.Errorf("could not read keywords: %v", err)
	}

	return keywords, nil
}

type config struct {
	DB struct {
		User     string `json:"user"`
		Pass     string `json:"pass"`
		Host     string `json:"host"`
		Database string `json:"database"`
	} `json:"db"`
}
