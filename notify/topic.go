package notify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/xerrors"
)

// GetTopics 指定したクライアントが購読しているトピックの一覧を取得する
func GetTopics(clientToken string) ([]string, error) {
	serverKey := os.Getenv("FIREBASE_SERVER_KEY")
	if serverKey == "" {
		return nil, xerrors.Errorf("Missing env 'FIREBASE_SERVER_KEY'")
	}

	url := fmt.Sprintf("https://iid.googleapis.com/iid/info/%v?details=true", url.PathEscape(clientToken))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("key=%v", serverKey))

	httpCli := &http.Client{}
	res, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, xerrors.Errorf("Can not get topics")
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	d := struct {
		Rel struct {
			Topics map[string]interface{} `json:"topics"`
		} `json:"rel"`
	}{}
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return nil, xerrors.Errorf("Can not unmarshl json: %w", err)
	}

	var topics []string
	for t := range d.Rel.Topics {
		topics = append(topics, t)
	}

	return topics, nil
}
