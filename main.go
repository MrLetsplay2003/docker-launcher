package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

	UpdateUID   bool
	User        string
	UIDVariable string

	UpdateGID   bool
	Group       string
	GIDVariable string

	RunBefore [][]string
}

const defaultVariableRegex = `\$\{([a-zA-Z0-9_-]+)\}`
const defaultUIDVariable = "UID"
const defaultGIDVariable = "GID"

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

func updateUID(config config) error {
	uidVariable := config.UIDVariable
	if len(uidVariable) == 0 {
		uidVariable = defaultUIDVariable
	}

	newUID := os.Getenv(uidVariable)
	if len(newUID) == 0 {
		log.Println("UID variable is not set. Not updating UID")
		return nil
	}

	log.Println("Updating UID of user", config.User)
	cmd := exec.Command("usermod", "-u", newUID, config.User)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func updateGID(config config) error {
	gidVariable := config.GIDVariable
	if len(gidVariable) == 0 {
		gidVariable = defaultGIDVariable
	}

	newGID := os.Getenv(gidVariable)
	if len(newGID) == 0 {
		log.Println("GID variable is not set. Not updating GID")
		return nil
	}

	log.Println("Updating GID of group", config.Group)
	cmd := exec.Command("groupmod", "-g", newGID, config.Group)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	force := flag.Bool("force", false, "Forcibly overwrite existing files (equivalent to setting forceOverwrite to true in the config)")
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("Missing command")
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("Failed to load config:", err)
	}

	config.ForceOverwrite = config.ForceOverwrite || *force

	err = filterFiles(*config)
	if err != nil {
		log.Fatalln(err)
	}

	if config.UpdateUID {
		err = updateUID(*config)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if config.UpdateGID {
		err = updateGID(*config)
		if err != nil {
			log.Fatalln(err)
		}
	}

	for _, command := range config.RunBefore {
		log.Println(">", strings.Join(command, " "))
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			log.Fatalln(err)
		}
	}

	command := flag.Args()
	log.Println(">", strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
