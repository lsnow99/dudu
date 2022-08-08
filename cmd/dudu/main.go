package main

import (
	"bufio"
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lsnow99/dudu/internal/rfsnotify"
	"gopkg.in/fsnotify.v1"
)

const tempDir = "./.dudu"

// Version of dudu
var Version string

// content holds our static web server content.
//go:embed templates
var content embed.FS

var srcDir, outputDir, resourceDir string
var httpPort int
var force bool

var buildCmd = flag.NewFlagSet("build", flag.ExitOnError)
var serveCmd = flag.NewFlagSet("serve", flag.ExitOnError)

func init() {
	buildCmd.StringVar(&outputDir, "output", "static", "Output directory")
	buildCmd.StringVar(&outputDir, "o", "static", "Output directory")
	buildCmd.StringVar(&srcDir, "source", "md", "Source directory for markdown")
	buildCmd.StringVar(&srcDir, "s", "md", "Source directory for markdown")
	buildCmd.StringVar(&resourceDir, "resources", "resources", "Resource directory")
	buildCmd.StringVar(&resourceDir, "r", "resources", "Resource directory")
	buildCmd.BoolVar(&force, "force", false, "Force regeneration of previously cached output files")
	buildCmd.BoolVar(&force, "f", false, "Force regeneration of previously cached output files")

	serveCmd.IntVar(&httpPort, "p", 8080, "Http port to listen on")
	serveCmd.IntVar(&httpPort, "port", 8080, "Http port to listen on")
	serveCmd.StringVar(&srcDir, "source", "md", "Source directory for markdown")
	serveCmd.StringVar(&srcDir, "s", "md", "Source directory for markdown")
	serveCmd.StringVar(&resourceDir, "resources", "resources", "Resource directory")
	serveCmd.StringVar(&resourceDir, "r", "resources", "Resource directory")
}

func main() {

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("no subcommand specified")
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "new":
		fmt.Printf("💎 Dudu %s project creator\n", Version)

		projectName := getOption("🟢 Project name", "personal-site")

		_, err := os.Stat(projectName)
		if os.IsNotExist(err) {
			if err := os.Mkdir(projectName, 0700); err != nil {
				log.Println("error:", err)
				break
			}
		} else {
			log.Printf("Folder %s already exists, exiting", projectName)
			break
		}

		populateProject(projectName, "templates")
		fmt.Printf("✨ Project created. Run `cd %s && dudu serve` to start working!\n", projectName)

	case "build":
		buildCmd.Parse(os.Args[2:])
		Generate(outputDir, false, force)

	case "serve":
		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			log.Println("error:", err)
			printUsage()
			break
		}

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}
		errCh := make(chan error)

		hub := newHub()
		go hub.run()

		wg.Add(2)
		go serveStatic(ctx, &wg, errCh, hub)
		go watchChanges(ctx, &wg, errCh, hub)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		select {
		case err := <-errCh:
			log.Println(err)
		case <-sigs:
		}
		cancel()
		wg.Wait()

		_ = os.RemoveAll(tempDir)

	case "help":
		printUsage()
	}
}

func watchChanges(ctx context.Context, wg *sync.WaitGroup, errCh chan error, hub *Hub) {
	defer wg.Done()

	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		select {
		case errCh <- err:
		default:
			return
		}
	}
	defer watcher.Close()

	go func() {
		<-ctx.Done()
		watcher.Close()
	}()

	fmt.Printf("Watching for changes in %s\n", srcDir)
	err = watcher.AddRecursive(srcDir)
	if err != nil {
		select {
		case errCh <- err:
		default:
			log.Println("error:", err)
			return
		}
	}

	if err := Generate(tempDir, true, false); err != nil {
		log.Println("error:", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				select {
				case errCh <- err:
				default:
					log.Println("error:", err)
					return
				}
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := Generate(tempDir, true, false); err != nil {
					log.Println("error:", err)
				}
				hub.broadcast <- []byte("update")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				select {
				case errCh <- err:
				default:
					log.Println("error:", err)
					return
				}
			}
			log.Println("error:", err)
		}
	}
}

func serveStatic(ctx context.Context, wg *sync.WaitGroup, errCh chan error, hub *Hub) {
	defer wg.Done()

	fmt.Printf("Listening on :%d\n", httpPort)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(tempDir)))
	mux.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hub, rw, r)
	})

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(httpPort),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		server.Shutdown(ctx)
		cancel()
	}()

	err := server.ListenAndServe()
	select {
	case errCh <- err:
	default:
	}
}

func populateProject(projectName, rootFolder string) {
	entries, err := content.ReadDir(rootFolder)
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		fullName := filepath.Join(rootFolder, entry.Name())
		if entry.IsDir() {
			populateProject(projectName, fullName)
		} else {
			// Create directory structure
			relName := filepath.Join(strings.Split(fullName, string(os.PathSeparator))[1:]...)
			outFile := filepath.Join(projectName, relName)
			outPath, _ := filepath.Split(outFile)

			if err := os.MkdirAll(outPath, 0700); err != nil {
				log.Println("error:", err)
				os.Exit(1)
			}

			// Read and copy the file
			data, err := content.ReadFile(fullName)
			if err != nil {
				log.Println("error:", err)
				os.Exit(1)
			}

			if err := os.WriteFile(outFile, data, 0700); err != nil {
				log.Println("error:", err)
				os.Exit(1)
			}
		}
	}
}

func printUsage() {
	fmt.Println("Available commands: new, build, serve")
	buildCmd.Usage()
	serveCmd.Usage()
}

func getOption(optionName, defaultValue string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s (%s): ", optionName, defaultValue)
	scanner.Scan()
	val := strings.TrimSpace(scanner.Text())
	if val == "" {
		return defaultValue
	}
	return val
}
