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

func writeXmlFile(xmlF string, ser XmlSerializer) (err error) {
	var xmlFile *os.File
	xmlFile, err = os.Create(xmlF)
	if err != nil {
		return err
	}
	defer func() {
		errClose := xmlFile.Close()
		if errClose != nil {
			err = errClose
		}
	}()

	xmlFileWriter := bufio.NewWriter(xmlFile)
	ser.WriteXml(xmlFileWriter)
	xmlFileWriter.Flush()
	return
}

func writeVmFile(vmF string, comp *Compiler) (err error) {
	var vmFile *os.File
	vmFile, err = os.Create(vmF)
	if err != nil {
		return err
	}
	defer func() {
		errClose := vmFile.Close()
		if errClose != nil {
			err = errClose
		}
	}()

	vmFileWriter := bufio.NewWriter(vmFile)
	vmFileWriter.WriteString(comp.String())
	vmFileWriter.Flush()
	return
}

func processJackFile(wg *sync.WaitGroup, errCh chan<- error, inF, xmlTkF, xmlTreeF string) {
	defer func() {
		wg.Done()
	}()

	isXml := xmlTkF != ""

	inFile, err := os.Open(inF)
	if err != nil {
		errCh <- err
		return
	}
	defer inFile.Close()

	tokenizer := NewTokenizer(bufio.NewReader(inFile))
	parser := NewPasreTree(tokenizer)
	rootTree, err := parser.Parse()
	if err != nil && !errors.Is(err, io.EOF) {
		errCh <- fmt.Errorf("File \"%s\" failed during parsing: %w", inF, err)
		return
	}

	if isXml {
		writeXmlFile(xmlTkF, tokenizer)
		writeXmlFile(xmlTreeF, parser)
	}

	compiler := NewCompiler()
	err = compiler.Run(rootTree)
	if err != nil {
		errCh <- fmt.Errorf("File \"%s\" failed during compilation: %w", inF, err)
	}
	vmFileName := getVmFileName(inF)
	fmt.Printf("Saving the vm file \"%s\"", vmFileName)
	writeVmFile(vmFileName, compiler)
}

func getTokenXmlFileName(inF string) string {
	fn := strings.TrimSuffix(inF, filepath.Ext(inF))
	return fn + "T.out.xml"
}

func getParserXmlFileName(inF string) string {
	fn := strings.TrimSuffix(inF, filepath.Ext(inF))
	return fn + ".out.xml"
}

func getVmFileName(inF string) string {
	fn := strings.TrimSuffix(inF, filepath.Ext(inF))
	return fn + ".vm"
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
		var xmlTkF, xmlTreeF string
		fmt.Printf("Reading the file \"%s\"\n", inF)
		if isXml {
			xmlTkF = getTokenXmlFileName(inF)
			xmlTreeF = getParserXmlFileName(inF)
			fmt.Printf("Saving results into \"%s\" and \"%s\"\n", xmlTkF, xmlTreeF)
		}
		wg.Add(1)
		go processJackFile(wg, errCh, inF, xmlTkF, xmlTreeF)
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
