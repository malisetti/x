package main

type item struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Deleted bool   `json:"deleted"`
	Dead    bool   `json:"dead"`

	DiscussLink string `json:"discussLink"`

	From     string `json:"from"`
	Priority int    `json:"priority"`

	Added    int    `json:"added"`
	RemoveAt int    `json:"removeAt"`
	Domain   string `json:"domain"`

	TweetID int64 `json:"tweetID"`
}
