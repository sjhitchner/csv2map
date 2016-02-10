package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DEFAULT_DELIMITER = ","

	USAGE_KEY_INDEX    = "Field index of key"
	USAGE_VALUE_INDEX  = "Field index of value"
	USAGE_MAP_NAME     = "Map name"
	USAGE_PACKAGE_NAME = "Go package name"
	USAGE_FILE_NAME    = "Path to CSV file"
	USAGE_DELIMITER    = "Field delimiter"
	USAGE_INSPECT      = "Parse and dump file to stdout"
	USAGE_HEADER       = "Ignore first header row"
)

var (
	keyIndex     int
	valueIndex   int
	mapName      string
	packageName  string
	fileName     string
	delimiter    string
	test         bool
	ignoreHeader bool
)

func init() {
	flag.IntVar(&keyIndex, "key", 0, USAGE_KEY_INDEX)
	flag.IntVar(&keyIndex, "k", 0, USAGE_KEY_INDEX)

	flag.IntVar(&valueIndex, "value", 0, USAGE_VALUE_INDEX)
	flag.IntVar(&valueIndex, "v", 0, USAGE_VALUE_INDEX)

	flag.StringVar(&mapName, "map", "", USAGE_MAP_NAME)
	flag.StringVar(&mapName, "m", "", USAGE_MAP_NAME)

	flag.StringVar(&packageName, "package", "", USAGE_PACKAGE_NAME)
	flag.StringVar(&packageName, "p", "", USAGE_PACKAGE_NAME)

	flag.StringVar(&fileName, "in", "", USAGE_FILE_NAME)
	flag.StringVar(&fileName, "i", "", USAGE_FILE_NAME)

	flag.StringVar(&delimiter, "delimiter", DEFAULT_DELIMITER, USAGE_DELIMITER)
	flag.StringVar(&delimiter, "d", DEFAULT_DELIMITER, USAGE_DELIMITER)

	flag.BoolVar(&ignoreHeader, "header", false, USAGE_HEADER)
	flag.BoolVar(&ignoreHeader, "h", false, USAGE_HEADER)

	flag.BoolVar(&test, "test", false, USAGE_INSPECT)
	flag.BoolVar(&test, "t", false, USAGE_INSPECT)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("csv2map -  ")

	flag.Parse()

	if err := process(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(-1)
		return
	}

	return
}

func process() error {
	if keyIndex < 1 {
		return fmt.Errorf("Invalid key index '%d'\n", keyIndex)
	}

	if valueIndex < 1 {
		return fmt.Errorf("Invalid value index '%d'\n", valueIndex)
	}

	if mapName == "" {
		return fmt.Errorf("Map name not provided")
	}

	if packageName == "" {
		return fmt.Errorf("Package name not provided")
	}

	reader, err := OpenFileOrStdInReader(fileName)
	if err != nil {
		return err
	}
	defer reader.Close()

	csvReader := CSVReader(reader, delimiter)
	if test {
		if err := dumpFile(csvReader); err != nil {
			return err
		}
		return nil
	}

	entries, err := processFile(csvReader, keyIndex, valueIndex, ignoreHeader)
	if err != nil {
		return err
	}

	outputFile, err := getOutputPath(fileName, mapName)
	if err != nil {
		return fmt.Errorf("Invalid output path: %s", err)
	}

	writer, err := OpenFileOrStdOutWriter(outputFile)
	if err != nil {
		return fmt.Errorf("Unable to open output file: %v", err)
	}
	defer writer.Close()

	if err := render(writer, packageName, mapName, entries); err != nil {
		return err
	}

	return nil
}

func dumpFile(reader *csv.Reader) error {
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		for i, field := range row {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(field)
		}
		fmt.Println()
	}
	return nil
}

func OpenFileOrStdInReader(fileIn string) (io.ReadCloser, error) {
	if fileIn == "" {
		log.Println("- Reading from STDIN -")
		return os.Stdin, nil
	}
	return os.Open(fileIn)
}

func OpenFileOrStdOutWriter(fileOut string) (io.WriteCloser, error) {
	if fileOut == "" {
		log.Println("- Writing to STDIN -")
		return os.Stdout, nil
	}
	return os.OpenFile(fileOut, os.O_WRONLY|os.O_CREATE, 0600)
}

func CSVReader(reader io.Reader, delimiter string) *csv.Reader {
	csvReader := csv.NewReader(reader)
	if delimiter != DEFAULT_DELIMITER {
		csvReader.Comma = GetSeparator(delimiter)
	}
	return csvReader
}

func GetSeparator(s string) rune {
	var sep string
	s = `'` + s + `'`
	sep, _ = strconv.Unquote(s)
	return ([]rune(sep))[0]
}

func getOutputPath(filePath string, mapName string) (string, error) {
	dir, _ := filepath.Split(filePath)
	return filepath.Join(dir, fmt.Sprintf("csv2map_%s.go", mapName)), nil
}
