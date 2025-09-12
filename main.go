package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"one_c_logs_converter/config"
	"one_c_logs_converter/log_event"
	"one_c_logs_converter/logger"
	"one_c_logs_converter/models"
	"one_c_logs_converter/promtail_writer"

	"github.com/fsnotify/fsnotify"
)

var version = "dev"

var (
	cfg        *config.Config
	configLock = new(sync.RWMutex)
	watcher    *fsnotify.Watcher
	configFilepath string

	processingFiles     = make(map[string]bool)
	processingFilesLock = new(sync.Mutex)
)

func main() {
	versionFlag := flag.Bool("version", false, "Print application version and exit")
	configFlag := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\nwritten by azheval <azheval@gmail.com>\n", version)
		os.Exit(0)
	}

	// Determine config file path
	if *configFlag != "" {
		configFilepath = *configFlag
	} else {
		exePath, err := os.Executable()
		if err != nil {
			println("Failed to get executable path: " + err.Error())
			os.Exit(1)
		}
		configFilepath = filepath.Join(filepath.Dir(exePath), "config.json")
	}

	// Load initial config
	var err error
	cfg, err = config.Load(configFilepath)
	if err != nil {
		println("Failed to load initial configuration: " + err.Error())
		os.Exit(1)
	}

	// Setup logger
	if err := logger.Setup(cfg.LogLevel, cfg.AppLogDir); err != nil {
		println("Failed to setup logger: " + err.Error())
		os.Exit(1)
	}

	slog.Info("starting one_c_logs_converter", "version", version)
	slog.Info("configuration loaded", "config", fmt.Sprintf("%+v", cfg))

	// Create and start watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create watcher", "error", err)
		os.Exit(1)
	}
	defer watcher.Close()

	updateWatcher()
	go initialScan()
	go eventLoop()

	<-make(chan struct{})
}

func reloadConfig() error {
	configLock.Lock()
	defer configLock.Unlock()

	newCfg, err := config.Load(configFilepath)
	if err != nil {
		return err
	}
	cfg = newCfg
	slog.Info("configuration reloaded", "config", fmt.Sprintf("%+v", cfg))
	// Re-setup logger with potentially new level
	if err := logger.Setup(cfg.LogLevel, cfg.AppLogDir); err != nil {
		slog.Error("failed to re-setup logger after config reload", "error", err)
	}
	return nil
}

func updateWatcher() {
	configLock.RLock()
	defer configLock.RUnlock()

	currentWatches := watcher.WatchList()
	newWatches := make(map[string]bool)
	newWatches[configFilepath] = true
	for _, project := range cfg.Projects {
		for _, dir := range project.LogDirs {
			if dir.Enabled {
				newWatches[dir.Path] = true
			}
		}
	}

	for _, w := range currentWatches {
		if !newWatches[w] {
			if err := watcher.Remove(w); err != nil {
				slog.Error("failed to remove watch", "path", w, "error", err)
			}
		}
	}

	for w := range newWatches {
		isWatching := false
		for _, currentWatch := range currentWatches {
			if currentWatch == w {
				isWatching = true
				break
			}
		}
		if !isWatching {
			if err := watcher.Add(w); err != nil {
				slog.Error("failed to add path to watcher", "path", w, "error", err)
			} else {
				slog.Info("watching path", "path", w)
			}
		}
	}
}

func initialScan() {
	configLock.RLock()
	projects := cfg.Projects
	configLock.RUnlock()

	for _, project := range projects {
		for _, dir := range project.LogDirs {
			if dir.Enabled {
				err := filepath.Walk(dir.Path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".xml") {
						tryProcessFile(path, project.Name, project.DeleteProcessed)
					}
					return nil
				})
				if err != nil {
					slog.Error("error during initial scan", "path", dir.Path, "error", err)
				}
			}
		}
	}
}

func eventLoop() {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Name == configFilepath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				slog.Info("config.json changed, reloading...")
				if err := reloadConfig(); err != nil {
					slog.Error("failed to reload config", "error", err)
				} else {
					updateWatcher()
				}
				continue
			}

			if (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) && strings.HasSuffix(event.Name, ".xml") {
				configLock.RLock()
				var projectName string
				var deleteProcessed bool
				var projectFound bool
				for _, p := range cfg.Projects {
					for _, d := range p.LogDirs {
						if strings.HasPrefix(event.Name, d.Path) {
							projectName = p.Name
							deleteProcessed = p.DeleteProcessed
							projectFound = true
							break
						}
					}
					if projectFound {
						break
					}
				}
				configLock.RUnlock()
				if projectFound {
					tryProcessFile(event.Name, projectName, deleteProcessed)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "error", err)
		}
	}
}

func tryProcessFile(filePath, projectName string, deleteProcessed bool) {
	processingFilesLock.Lock()
	if processingFiles[filePath] {
		processingFilesLock.Unlock()
		slog.Debug("file is already being processed, skipping", "name", filePath)
		return
	}
	processingFiles[filePath] = true
	processingFilesLock.Unlock()

	go func() {
		defer func() {
			processingFilesLock.Lock()
			delete(processingFiles, filePath)
			processingFilesLock.Unlock()
		}()
		processFile(filePath, projectName, deleteProcessed)
	}()
}

func processFile(filePath, projectName string, deleteProcessed bool) {
	// Create a structured logger that includes the project and file path in every message.
	logger := slog.With("project", projectName, "file", filePath)

	configLock.RLock()
	outpuDir := cfg.OutputDir
	configLock.RUnlock()

	logger.Info("processing file")

	var entries []models.Event
	var err error

	// Retry logic for reading the file
	for i := 0; i < 3; i++ {
		entries, err = log_event.Parse(filePath)
		if err == nil {
			break
		}
		logger.Warn("failed to parse file, retrying...", "error", err, "attempt", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Error("failed to parse file after multiple attempts", "error", err)
		return
	}

	if len(entries) > 0 {
		if err := promtail_writer.WriteEvents(outpuDir, projectName, entries); err != nil {
			logger.Error("failed to write events", "error", err)
			return
		}
		logger.Info("finished processing file", "events_count", len(entries))
	} else {
		logger.Info("no events found in file")
	}

	bakFilePath := filePath + ".bak"
	if err := os.Rename(filePath, bakFilePath); err != nil {
		logger.Error("failed to rename processed file", "error", err)
		return
	}
	logger.Info("renamed processed file", "to", bakFilePath)

	if deleteProcessed {
		if err := os.Remove(bakFilePath); err != nil {
			logger.Error("failed to delete processed file", "error", err)
		} else {
			logger.Info("deleted processed file")
		}
	}
}
