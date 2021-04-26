package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/lsnow99/dudu/internal/rfsnotify"
	"gopkg.in/fsnotify.v1"
)

const tempDir = "./.static"

var srcDir, outputDir, resourceDir string
var httpPort int

var buildCmd = flag.NewFlagSet("build", flag.ExitOnError)
var serveCmd = flag.NewFlagSet("serve", flag.ExitOnError)

func init() {
	buildCmd.StringVar(&outputDir, "output", "static", "Output directory")
	buildCmd.StringVar(&outputDir, "o", "static", "Output directory")
	buildCmd.StringVar(&srcDir, "source", "md", "Source directory for markdown")
	buildCmd.StringVar(&srcDir, "s", "md", "Source directory for markdown")
	buildCmd.StringVar(&resourceDir, "resources", "resources", "Resource directory")
	buildCmd.StringVar(&resourceDir, "r", "resources", "Resource directory")

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
	case "build":
		buildCmd.Parse(os.Args[2:])
		Generate(outputDir)

	case "serve":
		serveCmd.Parse(os.Args[2:])

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}
		errCh := make(chan error)

		wg.Add(2)
		go serveStatic(ctx, &wg, errCh)
		go watchChanges(ctx, &wg, errCh)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		select {
		case err := <-errCh:
			log.Println(err)
		case <-sigs:
		}
		cancel()
		wg.Wait()

		os.RemoveAll(tempDir)
	case "help":
		printUsage()
	}
}

func watchChanges(ctx context.Context, wg *sync.WaitGroup, errCh chan error) {
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

	if err := Generate(tempDir); err != nil {
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
				if err := Generate(tempDir); err != nil {
					log.Println("error:", err)
				}
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

func serveStatic(ctx context.Context, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()

	fmt.Printf("Listening on :%d\n", httpPort)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(httpPort),
		Handler: http.FileServer(http.Dir(tempDir)),
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

func printUsage() {
	fmt.Println("Available commands: build, serve")
	buildCmd.Usage()
	serveCmd.Usage()
}
