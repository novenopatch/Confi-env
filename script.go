package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"strconv"
)

type Config struct {
	Environments  map[string][]string `json:"environments"`
	CommonCommand string              `json:"common_command"`
}

func main() {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatal("Erreur lors de la récupération du chemin de l'exécutable:", err)
	}

	configPath := filepath.Join(filepath.Dir(executablePath), "config.json")

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal("Erreur lors du chargement du fichier de configuration:", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Veuillez spécifier un environnement (web, mobile, python, etc.).")
		os.Exit(1)
	}

	environment := os.Args[1]
	runCommonCommand := false
	if len(os.Args) > 2 {
		runCommonCommand, err = strconv.ParseBool(os.Args[2])
		if err != nil {
			log.Fatal("Le deuxième argument doit être un booléen.")
		}
	}
	if runCommonCommand  && config.CommonCommand != "" {
		err := runCommand(config.CommonCommand, nil)
		if err != nil {
			log.Fatalf("Erreur lors de l'exécution de la commande commune: %v", err)
		}
	}

	commands, ok := config.Environments[environment]
	if !ok {
		fmt.Println("Environnement non trouvé dans le fichier de configuration.")
		os.Exit(1)
	}

	stopCh := make(chan struct{})

	var wg sync.WaitGroup

	for _, command := range commands {
		wg.Add(1)
		go func(cmd string) {
			defer wg.Done()
			err := runCommand(cmd, stopCh)
			if err != nil {
				log.Printf("Erreur lors de l'exécution de la commande %s: %v", cmd, err)
			}
		}(command)
	}

	wg.Wait()

	close(stopCh)
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func runCommand(command string, stopCh <-chan struct{}) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command+" &")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		defer func() {
			if r := recover(); r != nil {
				
			}
		}()

		select {
		case <-stopCh:
			
			err := cmd.Process.Kill()
			if err != nil {
				log.Printf("Erreur lors de l'arrêt de la commande %s: %v", command, err)
			}
		default:
			
		}
	}()

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()

	if err != nil && err.Error() != "signal: killed" {
		return err
	}

	return nil
}
