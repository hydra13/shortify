package filestorage

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Record struct {
	ID          RecordID `json:"uuid"`
	ShortURL    string   `json:"short_url"`
	OriginalURL string   `json:"original_url"`
	UserID      string   `json:"user_id"`
}

type RecordID int

func (r RecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(r)))
}

func (r *RecordID) UnmarshalJSON(data []byte) error {
	cleanStr := strings.Trim(string(data), "\"")
	id, err := strconv.Atoi(cleanStr)

	if err == nil {
		*r = RecordID(id)
	}

	return err
}
