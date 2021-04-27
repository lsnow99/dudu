package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Generate(outputDir string, hotReload, force bool) error {
	srcFiles := make([]string, 0)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if matched, err := filepath.Match("*.md", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			srcFiles = append(srcFiles, filepath.Join(strings.Split(path, string(os.PathSeparator))[1:]...))
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, f := range srcFiles {
		iFilePath := filepath.Join(srcDir, f)
		oFilePath := filepath.Join(outputDir, strings.Replace(f, ".md", ".html", 1))
		oPath, _ := filepath.Split(oFilePath)

		if !force {
			ifi, iErr := os.Stat(iFilePath)
			ofi, oErr := os.Stat(oFilePath)
	
			if iErr == nil && oErr == nil && ifi.ModTime().Before(ofi.ModTime()) {
				continue
			}
		}

		// Create directory structure
		err := os.MkdirAll(oPath, 0700)
		if err != nil {
			return err
		}

		options := []string{
			"--standalone",
			"--css=/style.css",
			"--highlight-style=" + filepath.Join(resourceDir, "code-highlight.theme"),
			"--variable=lang:en",
			"--include-before-body=" + filepath.Join(resourceDir, "navbar.html"),
			"--include-after-body=" + filepath.Join(resourceDir, "footer.html"),
			"--template=" + filepath.Join(resourceDir, "template.html"),
		}

		if hotReload {
			options = append(options, "--include-in-header="+filepath.Join(resourceDir, "hotreload.html"))
		}

		options = append(options, iFilePath, "-o", oFilePath)

		// Run pandoc
		cmd := exec.Command("pandoc", options...)
		if err := cmd.Run(); err != nil {
			return err
		}

		log.Println("updated:", oFilePath)
	}

	staticFiles := make([]string, 0)
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if matched, err := filepath.Match("*.md", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			return nil
		}

		if matched, err := filepath.Match("*.swp", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			return nil
		}

		if matched, err := filepath.Match("*.swo", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			return nil
		}

		if matched, err := filepath.Match("*.bak", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			return nil
		}

		staticFiles = append(staticFiles, filepath.Join(strings.Split(path, string(os.PathSeparator))[1:]...))

		return nil
	})

	if err != nil {
		return err
	}

	for _, f := range staticFiles {
		iFilePath := filepath.Join(srcDir, f)
		oFilePath := filepath.Join(outputDir, f)
		oPath, _ := filepath.Split(oFilePath)

		ifi, iErr := os.Stat(iFilePath)
		ofi, oErr := os.Stat(oFilePath)

		if iErr == nil && oErr == nil && ifi.ModTime().Before(ofi.ModTime()) {
			continue
		}

		// Create directory structure
		err := os.MkdirAll(oPath, 0755)
		if err != nil {
			return err
		}

		// Copy the file
		in, err := os.Open(iFilePath)
		if err != nil {
			return err
		}

		out, err := os.Create(oFilePath)
		if err != nil {
			in.Close()
			return err
		}

		_, err = io.Copy(out, in)
		if err != nil {
			in.Close()
			out.Close()
			return err
		}

		out.Close()
		in.Close()
		log.Println("copied:", oFilePath)
	}

	return nil
}
