package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
)

type config struct {
	Files         []string
	VariableRegex string
	RegexGroup    int
}

var defaultVariableRegex = `\$\{([a-zA-Z0-9_-]+)\}`

func loadConfig(path string) (*config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config config
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func filterFile(path string, regex regexp.Regexp, regexGroup int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := string(data)

	text = regex.ReplaceAllStringFunc(text, func(s string) string {
		matches := regex.FindStringSubmatch(text)
		log.Println(matches)
		return os.Getenv(matches[regexGroup])
	})

	log.Println(text)

	return nil
}

func filterFiles(config config) error {
	var regexString string
	if len(config.VariableRegex) == 0 {
		regexString = defaultVariableRegex
	} else {
		regexString = config.VariableRegex
	}

	var regexGroup int
	if config.RegexGroup == 0 {
		regexGroup = 1
	} else {
		regexGroup = config.RegexGroup
	}

	log.Println("Regex is:", regexString)

	regex, err := regexp.Compile(regexString)
	if err != nil {
		return err
	}

	for _, file := range config.Files {
		log.Println("Filtering file:", file)
		err := filterFile(file, *regex, regexGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("Failed to load config:", err)
	}

	err = filterFiles(*config)
	if err != nil {
		log.Fatalln(err)
	}
}
