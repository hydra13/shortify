package filestorage

import (
	"encoding/json"
	"os"
)

func (fs *FileStorage) load() error {
	f, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		fs.log.Error().
			Err(err).
			Str("file_path", fs.filePath).
			Msg("can't open file storage")
		return err
	}
	f.Close()

	content, err := os.ReadFile(fs.filePath)
	if err != nil {
		fs.log.Error().
			Err(err).
			Str("file_path", fs.filePath).
			Msg("can't read file storage")
		return err
	}

	if len(content) == 0 {
		fs.log.Debug().
			Str("file_path", fs.filePath).
			Msg("empty file storage")
		return nil
	}

	records := make([]*Record, 0)
	err = json.Unmarshal(content, &records)
	if err != nil {
		fs.log.Error().
			Err(err).
			Str("file_path", fs.filePath).
			Msg("can't unmarshal file storage")
		return err
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.values = records
	for _, record := range records {
		fs.repository[record.ShortURL] = record
		if fs.freeID < record.ID+1 {
			fs.freeID = record.ID + 1
		}
	}

	return nil
}

func (fs *FileStorage) write() error {
	bytes, err := json.Marshal(fs.values)
	if err != nil {
		fs.log.Error().
			Err(err).
			Interface("repository", fs.repository).
			Msg("can't marshal file storage")
		return err
	}

	err = os.WriteFile(fs.filePath, bytes, 0o644)
	if err != nil {
		fs.log.Error().
			Err(err).
			Interface("repository", fs.repository).
			Str("file_path", fs.filePath).
			Msg("can't marshal file storage")
		return err
	}

	return nil
}
