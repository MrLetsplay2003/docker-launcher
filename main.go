package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
)

type file struct {
	Path       string
	TargetPath string
}

type config struct {
	Files          []file
	VariableRegex  string
	RegexGroup     int
	ForceOverwrite bool
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

func filterFile(file file, regex regexp.Regexp, regexGroup int) error {
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return err
	}

	stat, err := os.Stat(file.Path)
	if err != nil {
		return err
	}

	text := string(data)

	text = regex.ReplaceAllStringFunc(text, func(s string) string {
		matches := regex.FindStringSubmatch(text)
		return os.Getenv(matches[regexGroup])
	})

	err = os.WriteFile(file.TargetPath, []byte(text), stat.Mode())
	if err != nil {
		return err
	}

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
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
		if !config.ForceOverwrite {
			exists, err := fileExists(file.TargetPath)
			if err != nil {
				return err
			}

			if exists {
				log.Println("Skipping file:", file.Path, "->", file.TargetPath, "(already exists)")
				continue
			}
		}

		log.Println("Filtering file:", file.Path, "->", file.TargetPath)
		err := filterFile(file, *regex, regexGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	force := flag.Bool("force", false, "Forcibly overwrite existing files (equivalent to setting forceOverwrite to true in the config)")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("Failed to load config:", err)
	}

	config.ForceOverwrite = config.ForceOverwrite || *force

	err = filterFiles(*config)
	if err != nil {
		log.Fatalln(err)
	}
}
