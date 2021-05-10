package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// Exit codes
	argFail = iota + 1
	fsFail
	compFail
)

func parseArgs() (inPath string, isXml bool, err error) {
	flag.StringVar(&inPath, "in", "", "Input folder with *.jack files")
	flag.BoolVar(&isXml, "xml", false, "Generate output as xml files for testing purposes")
	flag.Parse()

	if inPath == "" {
		if inPath = flag.Arg(0); inPath == "" {
			err = errors.New("The input Path is not set")
			return
		}
	}
	return
}

// getJackFiles returns all jack files in the folder
func getJackFiles(dir string) ([]string, error) {
	dirPathInfo, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !dirPathInfo.IsDir() {
		return nil, fmt.Errorf("Input path \"%s\" is not a directory", dir)
	}

	var matched []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if m, err := filepath.Match("*.jack", filepath.Base(path)); err != nil {
			return err
		} else if m {
			matched = append(matched, path)
		}
		return nil
	})
	return matched, nil
}

func processJackFile(wg *sync.WaitGroup, errCh chan<- error, inF, xmlF string) {
	defer func() {
		wg.Done()
	}()

	isXml := xmlF != ""

	inFile, err := os.Open(inF)
	if err != nil {
		errCh <- err
		return
	}
	defer inFile.Close()

	var xmlFile *os.File

	tokenizer := NewTokenizer(bufio.NewReader(inFile))
	for {
		_, err := tokenizer.ReadToken()
		if err != nil && !errors.Is(err, io.EOF) {
			errCh <- fmt.Errorf("File \"%s\" failed during tokenizing: %w", inF, err)
			return
		} else if err != nil {
			break
		}
	}

	if isXml {
		xmlFile, err = os.Create(xmlF)
		if err != nil {
			errCh <- err
			return
		}
		defer func() {
			errClose := xmlFile.Close()
			if errClose != nil {
				errCh <- errClose
				return
			}
		}()

		xmlFileWriter := bufio.NewWriter(xmlFile)
		tokenizer.WriteXml(xmlFileWriter)
		xmlFileWriter.Flush()
	}
}

func getXmlFileName(inF string) string {
	fn := strings.TrimSuffix(inF, filepath.Ext(inF))
	return fn + "T.out.xml"
}

func gatherErrs(wg *sync.WaitGroup, errCh <-chan error) []error {
	errs := make([]error, 0)
	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
	}()

	for {
		select {
		case e := <-errCh:
			errs = append(errs, e)
		case <-done:
			return errs
		}
	}
}

func main() {
	inDirPath, isXml, err := parseArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Argument Error: %v", err))
		os.Exit(argFail)
	}

	inFiles, err := getJackFiles(inDirPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("File system error: %v", err))
		os.Exit(fsFail)
	}

	errCh := make(chan error)
	wg := &sync.WaitGroup{}

	for _, inF := range inFiles {
		var xmlF string
		fmt.Printf("Reading the file \"%s\"\n", inF)
		if isXml {
			xmlF = getXmlFileName(inF)
			fmt.Printf("Saving results into \"%s\"\n", xmlF)
		}
		wg.Add(1)
		go processJackFile(wg, errCh, inF, xmlF)
	}

	errs := gatherErrs(wg, errCh)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "Errors during compilation:")
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(compFail)
	}
}
