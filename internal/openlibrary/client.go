package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type OpenLibraryBook struct {
	Identifiers struct {
		Goodreads    []string `json:"goodreads"`
		Librarything []string `json:"librarything"`
	} `json:"identifiers"`
	Title   string `json:"title"`
	Authors []struct {
		Key string `json:"key"`
	} `json:"authors"`
	PublishDate   string   `json:"publish_date"`
	Publishers    []string `json:"publishers"`
	Covers        []int    `json:"covers"`
	Contributions []string `json:"contributions"`
	Languages     []struct {
		Key string `json:"key"`
	} `json:"languages"`
	SourceRecords []string `json:"source_records"`
	LocalID       []string `json:"local_id"`
	Type          struct {
		Key string `json:"key"`
	} `json:"type"`
	FirstSentence struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"first_sentence"`
	Key           string `json:"key"`
	NumberOfPages int    `json:"number_of_pages"`
	Works         []struct {
		Key string `json:"key"`
	} `json:"works"`
	Classifications struct{} `json:"classifications"`
	Ocaid           string   `json:"ocaid"`
	Isbn10          []string `json:"isbn_10"`
	Isbn13          []string `json:"isbn_13"`
	LatestRevision  int      `json:"latest_revision"`
	Revision        int      `json:"revision"`
	Created         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
}

type OpenLibraryAuthor struct {
	Bio       string `json:"bio"`
	RemoteIds struct {
		Viaf     string `json:"viaf"`
		Wikidata string `json:"wikidata"`
		Isni     string `json:"isni"`
	} `json:"remote_ids"`
	Photos         []int    `json:"photos"`
	Key            string   `json:"key"`
	AlternateNames []string `json:"alternate_names"`
	PersonalName   string   `json:"personal_name"`
	Links          []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Type  struct {
			Key string `json:"key"`
		} `json:"type"`
	} `json:"links"`
	SourceRecords []string `json:"source_records"`
	Type          struct {
		Key string `json:"key"`
	} `json:"type"`
	Name           string `json:"name"`
	BirthDate      string `json:"birth_date"`
	DeathDate      string `json:"death_date"`
	LatestRevision int    `json:"latest_revision"`
	Revision       int    `json:"revision"`
	Created        struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
}

func FindBookByISBN(ctx context.Context, logger *slog.Logger, isbn string) (OpenLibraryBook, []OpenLibraryAuthor, error) {
	var book OpenLibraryBook
	if err := findByKey(fmt.Sprintf("/isbn/%s", isbn), &book); err != nil {
		return OpenLibraryBook{}, nil, err
	}

	var authors []OpenLibraryAuthor

	if len(book.Authors) > 0 {
		authors = make([]OpenLibraryAuthor, 0, len(book.Authors))
		for _, item := range book.Authors {
			if len(item.Key) > 0 {
				var author OpenLibraryAuthor
				if err := findByKey(item.Key, &author); err != nil {
					logger.ErrorContext(ctx, "failed to get author", slog.String("author", item.Key), slog.Any("error", err))
					continue
				}
				authors = append(authors, author)
			}
		}
	}

	return book, authors, nil
}

func findByKey(key string, result any) error {
	url := fmt.Sprintf("https://openlibrary.org/%s.json", key)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch book data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
