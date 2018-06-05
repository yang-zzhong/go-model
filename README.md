## go model and repo

### sample
```go

import (
    model "github.com/yang-zzhong/go-model"
    "time"
)

// define a model
type User struct {
    Id          string      `db:"id char(36) pk"`
    Account     string      `db:"account varchar(36) uk"`
    Name        string      `db:"name varchar(36)"`
    Birthday    time.Time   `db:"birthday datetime"`
    *model.Base
}
// define table of user
func (user *User) TableName() string {
    return "user"
}
// user constructor
func NewUser() *User {
    user := NewModel(new(User)).(*User)
    user.Id = helpers.RandString(32)
    user.DeclareMany("books", new(Book), map[string]string {
        "id": "author_id",
    })

    return user
}

type Book struct {
    Id         string       `db:"id char(36) pk"`
    Name       string       `db:"name varchar(32)"`
    AuthorId   string       `db:"author_id char(36)"`
    PublishedAt time.Time   `db:"published_at datetime"`
    *model.Base
}
// define book table name
func (book *Book) TableName() {
    return "books"
}
// define book constructor
func NewBook() *Book {
    book := NewModel(new(Book)).(*Book)
    book.Id = helpers.RandString(32)
    book.DeclareOne("author", new(User), map[string]string{
        "author_id": "id",
    })

    return book
}

// create user
user := NewUser()
user.Fill(map[string]interface{}{
    "name": "Mr. Bob",
    "age": 15,
    "account": "Mr_Bob",
    "birthday": time.Now(),
})
if err := user.Create(); err != nil {
    panic(err)
}

// create book
book := NewBook()
book.Name = "one two three"
book.AuthorId = user.Id
book.PublishedAt = time.Now()
if err := book.Create(); err != nil {
    panic(err)
}

// get many book
if books, err := user.Many("books"); err != nil {
    panic(err)
} else {
    for book_id, m := range books {
        book := m.(*Book)
        // handle book
    }
}

// get one author
if m, err := user.One("author"); err != nil {
    panic(err)
} else {
    author := m.(*User)
    // handle author
}

// fetch user with many book
if models, err := user.Repo().WithMany("books").Fetch(); err != nil {
    panic(err)
} else {
    for id, m := range models {
        user := m.(*User)
        if books, err := user.Many("books"); err == nil {
            // handle books
        } else {
            panic(err)
        }
    }
}

// fetch book with one author
if models, err := book.Repo().WithOne("author").Fetch(); err != nil {
    panic(err)
} else {
    for id, m := range models {
        book := m.(*Book)
        if author, err := book.One("author"); err == nil {
            // handle user
        } else {
            panic(err)
        }
    }
}
```
