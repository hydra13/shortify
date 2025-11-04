package filestorage

import (
	"encoding/json"
	"os"
)

func (fs *FileStorage) load() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	f, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	f.Close()

	content, err := os.ReadFile(fs.filePath)
	if err != nil {
		panic(err)
	}

	if len(content) == 0 {
		return
	}

	records := make([]*Record, 0)
	err = json.Unmarshal(content, &records)
	if err != nil {
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

func (fs *FileStorage) write() {
	bytes, err := json.Marshal(fs.values)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fs.filePath, bytes, 0o644)
	if err != nil {
		panic(err)
	}
}
