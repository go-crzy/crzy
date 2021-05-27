package pkg

import (
	"bufio"
	"os"
	"sync"
)

type file struct {
	sync.Mutex
	filename string
}

func (f *file) Write(p []byte) (int, error) {
	f.Lock()
	defer f.Unlock()
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(p)
}

func (f *file) ReadLines(offset, limit int) ([]string, error) {
	f.Lock()
	defer f.Unlock()
	file, err := os.Open(f.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	output := []string{}
	line := -1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line++
		if line < offset {
			continue
		}
		if line >= offset+limit {
			break
		}
		line := scanner.Text()
		output = append(output, line)
	}
	return output, nil
}
