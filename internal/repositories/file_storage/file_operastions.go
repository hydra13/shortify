package filestorage

import (
	"context"
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

	ctx := context.Background()

	for _, record := range records {
		err := fs.inMemory.Add(ctx, record.ShortURL, record.OriginalURL, record.UserID)
		if err != nil {
			fs.log.Error().
				Err(err).
				Interface("record", record).
				Str("file_path", fs.filePath).
				Msg("can't add record to storage")
			return err
		}
	}

	return nil
}

func (fs *FileStorage) write(ctx context.Context) error {
	values, err := fs.inMemory.GetAll(ctx)
	if err != nil {
		fs.log.Error().
			Err(err).
			Msg("can't get all records from storage")
		return err
	}

	records := make([]*Record, 0)
	for key, value := range values {
		records = append(records, &Record{
			ID:          RecordID(len(records) + 1),
			ShortURL:    key,
			OriginalURL: value,
		})
	}
	bytes, err := json.Marshal(records)
	if err != nil {
		fs.log.Error().
			Err(err).
			Interface("records", records).
			Msg("can't marshal file storage")
		return err
	}

	err = os.WriteFile(fs.filePath, bytes, 0o644)
	if err != nil {
		fs.log.Error().
			Err(err).
			Interface("records", records).
			Str("file_path", fs.filePath).
			Msg("can't marshal file storage")
		return err
	}

	return nil
}
