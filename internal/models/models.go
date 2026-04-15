package models

// User is the basic account data used in handlers and templates.
type User struct {
	ID       int
	Email    string
	Username string
}

// Category is used to group posts and filter the feed.
type Category struct {
	ID   int
	Name string
}

// Post holds the data shown on the feed and post page.
type Post struct {
	ID         int
	Title      string
	Content    string
	Author     string
	UserID     int
	CreatedAt  int64
	Likes      int
	Dislikes   int
	Categories []Category
}

// Comment holds one reply under a post.
type Comment struct {
	ID        int
	Content   string
	Author    string
	UserID    int
	CreatedAt int64
	Likes     int
	Dislikes  int
}

// TemplateData is the shared view model passed into HTML templates.
type TemplateData struct {
	User       *User
	Categories []Category
	Posts      []Post
	Post       *Post
	Comments   []Comment
	Filter     string
	CategoryID int
	FormError  string
	Info       string
	Status     int
	HasNext    bool
	HasPrev    bool
	NextPage   int
	PrevPage   int
	Page       int
	PageQuery  string
}
