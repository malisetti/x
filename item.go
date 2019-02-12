package main

type item struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Deleted     bool   `json:"deleted"`
	Dead        bool   `json:"dead"`
	DiscussLink string `json:"discussLink"`
	Added       int    `json:"added"`
	Domain      string `json:"domain"`

	Descriprion string   `json:"description"`
	Images      []string `json:"images"`
	TweetID     int64    `json:"tweetID"`
}
