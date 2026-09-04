package main

import (
	"fmt"
	"io/ioutil"
	"log"
	
	"gopkg.in/yaml.v3"
	
	"github.com/1jehuang/backtest/internal/strategy/parser"
)

func main() {
	// Load strategy
	data, err := ioutil.ReadFile("/tmp/simple_test_strategy.yaml")
	if err != nil {
		log.Fatal(err)
	}
	
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Raw YAML entry section:")
	if strategy, ok := raw["strategy"].(map[string]interface{}); ok {
		if entry, ok := strategy["entry"].(map[string]interface{}); ok {
			if long, ok := entry["long"].(map[string]interface{}); ok {
				fmt.Printf("Long entry: %+v\n\n", long)
			}
		}
	}
	
	// Parse strategy
	p := parser.NewParser()
	strategy, err := p.Parse(data)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	
	fmt.Println("Parsed strategy:")
	fmt.Printf("Name: %s\n", strategy.Name)
	fmt.Printf("Indicators count: %d\n", len(strategy.Indicators))
	fmt.Printf("Entry.Long: %+v\n", strategy.Entry.Long)
	
	if strategy.Entry.Long != nil {
		fmt.Printf("Entry.Long.Type: %s\n", strategy.Entry.Long.Type)
		fmt.Printf("Entry.Long.Function: %s\n", strategy.Entry.Long.Function)
		fmt.Printf("Entry.Long.Args: %v\n", strategy.Entry.Long.Args)
	}
}
