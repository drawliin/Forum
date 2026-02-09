package app

type User struct {
    ID       int
    Email    string
    Username string
}

type Category struct {
    ID   int
    Name string
}

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

type Comment struct {
    ID        int
    Content   string
    Author    string
    UserID    int
    CreatedAt int64
    Likes     int
    Dislikes  int
}

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
}
