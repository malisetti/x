package app

// Item is HN item
type Item struct {
	//ID is item's unique id.
	ID int `json:"id"`
	//By is the username of the item's author.
	By string `json:"by"`
	//Title of the story, poll or job.
	Title string `json:"title"`
	//URL of the story.
	URL string `json:"url"`
	//Deleted denotes if the item is deleted.
	Deleted bool `json:"deleted"`
	//Dead denotes if the item is dead.
	Dead bool `json:"dead"`
	//DiscussLink is where the discussion on this item goes.
	DiscussLink string `json:"discussLink"`
	//Added is the time of addition of item.
	Added int `json:"added"`
	//Domain is the domain from where the story is delivered.
	Domain string `json:"domain"`

	//EncryptedURL is the URL encrypted using key from config.
	EncryptedURL string `json:"encryptedURL"`
	//EncryptedDiscussLink is the URL encrypted using key from config.
	EncryptedDiscussLink string `json:"encryptedDiscussLink"`

	//Description of the item.
	Description string `json:"description"`
	//Source of the item
	Source string `json:"source"`
}
