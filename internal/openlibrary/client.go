package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	Key             string   `json:"key"`
	NumberOfPages   int      `json:"number_of_pages"`
	Classifications struct{} `json:"classifications"`
	Ocaid           string   `json:"ocaid"`
	Isbn10          []string `json:"isbn_10"`
	Isbn13          []string `json:"isbn_13"`
	LatestRevision  int      `json:"latest_revision"`
	Revision        int      `json:"revision"`
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
	Name           string   `json:"name"`
	BirthDate      string   `json:"birth_date"`
	DeathDate      string   `json:"death_date"`
	LatestRevision int      `json:"latest_revision"`
	Revision       int      `json:"revision"`
}

type Book struct {
	Title           string
	Author          string
	ISBN            string
	PublicationYear int
	CoverImage      []byte
}

func FindBookByISBN(ctx context.Context, logger *slog.Logger, isbn string) (Book, error) {
	var book OpenLibraryBook
	if err := findByKey(fmt.Sprintf("/isbn/%s", isbn), &book); err != nil {
		return Book{}, err
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

	return mapToBook(book, authors)
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

func mapToBook(book OpenLibraryBook, authors []OpenLibraryAuthor) (Book, error) {
	authorName := ""
	if len(authors) > 0 {
		authorNames := make([]string, 0, len(authors))
		for _, author := range authors {
			authorNames = append(authorNames, author.Name)
		}
		authorName = strings.Join(authorNames, ", ")
	}

	var coverImage []byte
	if len(book.Covers) > 0 {
		coverUrl := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", book.Covers[0])
		resp, err := http.Get(coverUrl)
		if err != nil {
			return Book{}, fmt.Errorf("failed to download cover image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			coverImage, err = io.ReadAll(resp.Body)
			if err != nil {
				return Book{}, fmt.Errorf("failed to read cover image data: %w", err)
			}
		}
	}

	year := 0
	// Regular expression to match a 4-digit year
	re := regexp.MustCompile(`\b\d{4}\b`)

	// Find the first match
	if yearString := re.FindString(book.PublishDate); yearString != "" {
		var err error
		if year, err = strconv.Atoi(yearString); err != nil {
			return Book{}, fmt.Errorf("cannot get year from publish date '%s': %w", book.PublishDate, err)
		}
	}

	// The return type only supports one ISBN
	isbn := ""
	if len(book.Isbn13) > 0 {
		isbn = book.Isbn13[0]
	} else if len(book.Isbn10) > 0 {
		isbn = book.Isbn10[0]
	}

	return Book{
		Title:           book.Title,
		Author:          authorName,
		ISBN:            isbn,
		PublicationYear: year,
		CoverImage:      coverImage,
	}, nil
}
