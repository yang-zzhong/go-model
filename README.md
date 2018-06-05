## go model and repo

### sample
```go

import (
    . "github.com/yang-zzhong/go-querybuilder"
    model "github.com/yang-zzhong/go-model"
    helpers "github.com/yang-zzhong/go-helpers"
    "database/sql"
)

// define a model
type User struct {
    Id          string      `db:"id char(36) pk"`
    Account     string      `db:"account varchar(36) uk"`
    Name        string      `db:"name varchar(36)"`
    Birthday    time.Time   `db:"birthday datetime"`
    *model.Base
}

func (user *User) PK() string {
    return "id"
}

func (user *User) TableName() string {
    return "user"
}

func (user *User) Many(name string) (map[interface{]]interface{}, error) {
    return model.Many(user.Base, user, name)
}

func NewUser() *User {
    user := new(User)
    user.Id = helpers.RandString(32)
    user.Base = model.NewBase(user)
    book := new(Book)
    book.Base = model.NewBase(book)
    user.DeclareMany("books", book, map[string]string {
        "id": "user_id",
    })

    return user
}

type Book struct {
    Id         string       `db:"id char(36) pk"`
    Name       string       `db:"name varchar(32)"`
    AuthorId   string       `db:"author_id char(36)"`
    PublishedAt time.Time   `db:"published_at datetime"`
}

func (book *Book) PK() {
    return "id"
}

func (book *Book) TableName() {
    return "books"
}

func (book *Book) One(name string) (interface{}, error) {
    return model.One(book.Base, book, name)
}

func NewBook() *Book {
    book := new(Book)
    book.Id = helpers.RandString(32)
    book.Base = NewBase(book)

    user := new(User)
    user.Base = NewBase(user)

    book.DeclareOne("author", user, map[string]string{
        "author_id": "id",
    })

    return book
}

// create user
user := NewUser()
user.Name = "Mr. Bob"
user.Age = 15
user.Account = "Mr_Bob"
user.Birthday = time.Now()
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
        book := m.(Book)
    }
}

// get one author
if m, err := user.One("author"); err != nil {
    panic(err)
} else {
    user := m.(User)
}

// fetch user with many book
if models, err := user.Repo().WithMany("books").Fetch(); err != nil {
    panic(err)
} else {
    for id, model := range models {
        user := model.(User)
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
    for id, model := range models {
        book := model.(Book)
        if user, err := user.One("author"); err == nil {
            // handle user
        } else {
            panic(err)
        }
    }
}
```
