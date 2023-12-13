package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const timeFormat = "Jan/02 @15:04"

var (
	ErrConnection      = errors.New("Connection error")
	ErrNotFound        = errors.New("Not found")
	ErrInvalidResponse = errors.New("Invalid server response")
	ErrInvalid         = errors.New("Invalid data")
	ErrNotNumber       = errors.New("Not a number")
)

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type response struct {
	Results      []item `json:"results"`
	Date         int    `json:"date"`
	TotalResults int    `json:"total_results"`
}

func newClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 20,
	}
	c := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	return c
}

func getItems(url string) ([]item, error) {
	r, err := newClient().Get(url)
	if err != nil {
		// log.Println("This is where we're erroring out!")
		// Error: Connection error: Get "http://localhost:8080/todo": read tcp [::1]:61138->[::1]:10011: read: connection reset by peer
		return nil, fmt.Errorf("%w: %s", ErrConnection, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("Cannot read body: %w", err)
		}
		err = ErrInvalidResponse
		if r.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}
		return nil, fmt.Errorf("%w: %s", err, msg)
	}

	var resp response

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.TotalResults == 0 {
		return nil, fmt.Errorf("%w: No results found", ErrNotFound)
	}

	return resp.Results, nil
}

func getAll(apiRoot string) ([]item, error) {
	u := fmt.Sprintf("%s/todo", apiRoot)

	return getItems(u)
}

func getOne(apiRoot string, id int) (item, error) {
	u := fmt.Sprintf("%s/todo/%d", apiRoot, id)

	items, err := getItems(u)
	if err != nil {
		return item{}, err
	}
	if len(items) != 1 {
		return item{}, fmt.Errorf("%w: Invalid results", ErrInvalid)
	}

	return items[0], nil
}
