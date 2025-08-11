package main

import (
	"errors"
	"flag"
	"strings"
)

type config struct {
	inputFile     string
	outputFile    string
	testCasesFile string
	requestTypes  []string
}

func (cfg config) validate() error {
	switch {
	case cfg.inputFile == "":
		return errors.New("input file is required")
	case cfg.outputFile == "":
		return errors.New("output file is required")
	case cfg.testCasesFile == "":
		return errors.New("test cases file is required")
	}
	return nil
}

func initConfig() (config, error) {
	var (
		cfg      config
		reqTypes string
	)

	flag.StringVar(&cfg.inputFile, "input", "", "Input Go file to parse")
	flag.StringVar(&cfg.outputFile, "output", "", "Output test file")
	flag.StringVar(&cfg.testCasesFile, "testcases", "", "JSON file containing test cases (defaults to <input>_testcases.json)")
	flag.StringVar(&reqTypes, "request-type", "", "Supported request types")
	flag.Parse()

	for _, rt := range strings.Split(reqTypes, ",") {
		cfg.requestTypes = append(cfg.requestTypes, strings.TrimSpace(rt))
	}

	return cfg, cfg.validate()
}
