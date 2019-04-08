package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var Input string
var Output string
var Interval time.Duration
var Version string

var rootCmd = &cobra.Command{
	Use:   "bak",
	Short: "Target a file or directory to be backed up regularly",
	Long:  `A utility to watch a file or directory for changes and back up the file(s) to a separate location at specified intervals.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get input
		input, err := filepath.Abs(Input)
		if err != nil {
			log.Fatalf("Unable to parse input file path: %s", err)
		}
		// get output
		output, err := filepath.Abs(Output)
		if err != nil {
			log.Fatalf("Unable to parse output file path: %s", err)
		}

		fmt.Printf("Watching %s, backing up to %s every %s\n", input, output, Interval)

		w := watcher.New()
		w.FilterOps(watcher.Write, watcher.Create)

		go func() {
			for {
				select {
				case event := <-w.Event:
					if event.IsDir() {
						continue
					}
					from := event.Path
					relPath := strings.TrimPrefix(from, input)
					basePath := filepath.Dir(filepath.Join(output, relPath))
					if err := os.MkdirAll(basePath, 0755); err != nil {
						log.Fatalf("Unable to create output directory: %s", err)
					}
					to := filepath.Join(basePath, time.Now().Format("2006-01-02_15.04_")+filepath.Base(from))
					log.Printf("Copying from %s to %s\n", from, to)
					if _, err := copyFile(from, to); err != nil {
						log.Fatalf("Error copying file: %s", err)
					}
				case err := <-w.Error:
					log.Printf("Error watching file: %s", err)
				case <-w.Closed:
					return
				}
			}
		}()

		if err := w.AddRecursive(input); err != nil {
			log.Fatalf("Unable to watch input: %s", err)
		}

		go func() {
			w.Wait()
			for p, f := range w.WatchedFiles() {
				if !f.IsDir() {
					w.Event <- watcher.Event{Op: watcher.Create, Path: p, FileInfo: f}
				}
			}
		}()

		if err := w.Start(Interval); err != nil {
			log.Fatalf("Error starting file watcher: %s", err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&Input, "input", "", "the path to the file or directory to watch")
	rootCmd.MarkFlagRequired("input")
	rootCmd.PersistentFlags().StringVar(&Output, "output", "", "the path to the directory where files should be backed up to")
	rootCmd.MarkFlagRequired("output")
	rootCmd.PersistentFlags().DurationVar(&Interval, "interval", time.Minute*5, "the interval to back up changed files")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
