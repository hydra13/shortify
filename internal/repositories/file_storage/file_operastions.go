package filestorage

import (
	"encoding/json"
	"io"
	"os"
)

func (fs *FileStorage) load() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	f, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	for {
		records := make([]*Record, 0)
		if err := decoder.Decode(&records); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		fs.values = records
		for _, record := range records {
			fs.repository[record.ShortURL] = record
			if fs.freeID < record.ID+1 {
				fs.freeID = record.ID + 1
			}

		}
	}

	for _, v := range fs.repository {
		fs.values = append(fs.values, v)
	}
}

func (fs *FileStorage) write() {
	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(fs.values)
	if err != nil {
		panic(err)
	}
}
