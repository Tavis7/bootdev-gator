package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return RSSFeed{}, err
	}
	if res.StatusCode >= 300 {
		return RSSFeed{}, fmt.Errorf("Unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	bodyText, err := io.ReadAll(res.Body)
	if err != nil {
		return RSSFeed{}, err
	}

	result := RSSFeed{}

	err = xml.Unmarshal(bodyText, &result)
	if err != nil {
		return RSSFeed{}, err
	}

	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)

	for i := 0; i < len(result.Channel.Item); i++ {
		result.Channel.Item[i].Title =
			html.UnescapeString(result.Channel.Item[i].Title)
		result.Channel.Item[i].Description =
			html.UnescapeString(result.Channel.Item[i].Description)
	}

	return result, nil
}
