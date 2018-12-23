## go model and repo

### sample
```go

import (
    model "github.com/yang-zzhong/go-model"
    query "github.com/yang-zzhong/go-querybuilder"
    "database/sql"
    _ "github.com/lib/pq"
    "time"
)

// init
func init() {
    db, err := sql.Open("postgres", "postgres://test:123456@host.com/database")
    if err != nil {
        panic(err)
    }
    model.RegisterDefaultDB(db, &query.PgsqlModifier{})
}

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

func (user *User) Prepare() {
    user.DeclareMany("books", new(Book), map[string]string {
        "id": "author_id",
    })
}

// user constructor
func NewUser() *User {
    return model.NewModel(new(User)).(*User)
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

func (book *Book) Prepare() {
    book.DeclareOne("author", new(User), map[string]string{
        "author_id": "id",
    })
}

// define book constructor
func NewBook() *Book {
    return model.NewModel(new(Book)).(*Book)
}

// create user
user := NewUser()
user.Fill(map[string]interface{}{
    "name": "Mr. Bob",
    "age": 15,
    "account": "Mr_Bob",
    "birthday": time.Now(),
})
if err := user.Save(); err != nil {
    panic(err)
}

// create book
book := NewBook()
book.Name = "one two three"
book.AuthorId = user.Id
book.PublishedAt = time.Now()
if err := book.Save(); err != nil {
    panic(err)
}

// get many book
books := user.MustMany("books")
for _, m := range books {
    book := m.(*Book)
    // handle book
}

// get one author
m := user.MustOne("author")
author := m.(*User)
// handle author

// fetch user with many book
models := user.Repo().With("books").MustFetch()
for _, m := range models {
    user := m.(*User)
    books := user.MustMany("books")
}

// fetch book with one author
models := book.Repo().With("author").MustFetch()
for _, m := range models {
    book := m.(*Book)
    author := book.MustOne("author")
}
```

[doc](https://booblogger.com/go-model)
