package main

type item struct {
	//from HN
	ID          int    `json:"id"`      //The item's unique id.
	By          string `json:"by"`      //The username of the item's author.
	Textx       string `json:"text"`    //The comment, story or poll text. HTML.
	Title       string `json:"title"`   //The title of the story, poll or job.
	URL         string `json:"url"`     //The URL of the story.
	Deleted     bool   `json:"deleted"` //true if the item is deleted.
	Dead        bool   `json:"dead"`    //true if the item is dead.
	DiscussLink string `json:"discussLink"`
	Added       int    `json:"added"`
	Domain      string `json:"domain"`

	EncryptedURL         string `json:"-"`
	EncryptedDiscussLink string `json:"-"`

	Description string `json:"description"`
	TweetID     int64  `json:"-"`
}
