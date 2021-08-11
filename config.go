package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (d *Ingest) readEnvironmentFile() (map[string]string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	// log.Println("path=", path)
	f, err := os.Open(path + "/config.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := make(map[string]string)
	s := bufio.NewScanner(f)
	//read each line and separate the key and value which is separated by a =
	// ignore lines that start with a #
	// ignore lines that are empty
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\n") {
			continue
		}
		if strings.HasPrefix(line, "\r") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid environment file format")
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}
