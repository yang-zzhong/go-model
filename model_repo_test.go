package model

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/yang-zzhong/go-querybuilder"
	"log"
	. "testing"
	"time"
)

type User struct {
	Id        string    `db:"id varchar(128) pk"`
	Name      string    `db:"name varchar(32) uk"`
	Age       int       `db:"age int"`
	Level     int       `db:"level int"`
	Optional  string    `db:"optional varchar(256) nil"`
	CreatedAt time.Time `db:"created_at datetime"`
	*BaseModel
}

type Book struct {
	Id     string `db:"id varchar(128) pk"`
	UserId string `db:"user_id varchar(128)"`
	Name   string `db:"name varchar(128)"`
	*BaseModel
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) PK() string {
	return "id"
}

func (b *Book) TableName() string {
	return "book"
}

func (b *Book) PK() string {
	return "id"
}

func (u *User) One(name string) (interface{}, error) {
	return One(u.BaseModel, u, name)
}

func (u *User) Many(name string) (interface{}, error) {
	return Many(u.BaseModel, u, name)
}

func (b *Book) Many(name string) (map[interface{}]interface{}, error) {
	return Many(b.BaseModel, b, name)
}

func (b *Book) One(name string) (interface{}, error) {
	return One(b.BaseModel, b, name)
}

func NewUser() *User {
	user := new(User)
	user.BaseModel = NewBaseModel()
	user.DeclareMany("books", new(Book), map[string]string{
		"id": "user_id",
	})
	return user
}

func NewBook() *Book {
	book := new(Book)
	book.BaseModel = NewBaseModel()
	book.DeclareOne("author", new(User), map[string]string{
		"user_id": "id",
	})
	return book
}

func init() {
	var err error
	var con *sql.DB
	con, err = sql.Open("mysql", "root:young159357789@/test_go?parseTime=true")
	if err != nil {
		panic(err)
	}
	Config(con, &MysqlModifier{})
}

func TestWithMany(t *T) {
	var user *User
	var err error
	ur := createUserRepo()
	br := createBookRepo()
	if user, err = insertUser(ur); err != nil {
		clearRepo(ur)
		clearRepo(br)
		panic(err)
	}
	if _, err := insertBook(br); err != nil {
		clearRepo(ur)
		clearRepo(br)
		panic(err)
	}
	log.Print("books")
	log.Print(user.Many("books"))
	clearRepo(ur)
	clearRepo(br)
}

func TestWithOne(t *T) {
	var book *Book
	var err error
	ur := createUserRepo()
	br := createBookRepo()
	if _, err = insertUser(ur); err != nil {
		clearRepo(ur)
		clearRepo(br)
		panic(err)
	}
	if book, err = insertBook(br); err != nil {
		clearRepo(ur)
		clearRepo(br)
		panic(err)
	}
	log.Print("author")
	log.Print(book.One("author"))
	clearRepo(ur)
	clearRepo(br)
}

func TestCreate(t *T) {
	repo := createUserRepo()
	if _, err := insertUser(repo); err != nil {
		clearRepo(repo)
		panic(err)
	}
	clearRepo(repo)
}

func TestFetch(t *T) {
	repo := createUserRepo()
	if _, err := insertUser(repo); err != nil {
		clearRepo(repo)
		panic(err)
	}
	if rows, err := repo.Fetch(); err != nil {
		clearRepo(repo)
		panic(err)
	} else {
		log.Print(rows)
	}
	clearRepo(repo)
}

func TestFind(t *T) {
	repo := createUserRepo()
	if _, err := insertUser(repo); err != nil {
		clearRepo(repo)
		panic(err)
	}
	if model, err := repo.Find("1"); err != nil {
		clearRepo(repo)
	} else {
		log.Print(model)
	}
	clearRepo(repo)
}

func insertUser(repo *Repo) (*User, error) {
	user := NewUser()
	user.Id = "1"
	user.Name = "yang-zhong"
	user.Age = 17
	user.Level = 1
	user.CreatedAt = time.Now()

	return user, repo.Create(user)
}

func insertBook(repo *Repo) (*Book, error) {
	book := NewBook()
	book.UserId = "1"
	book.Name = "hello world"
	book.Id = "1"

	return book, repo.Create(book)
}

func createUserRepo() *Repo {
	var repo *Repo
	var err error
	if repo, err = NewRepo(NewUser()); err != nil {
		panic(err)
	}
	if err = repo.CreateRepo(); err != nil {
		panic(err)
	}
	return repo
}

func createBookRepo() *Repo {
	var repo *Repo
	var err error
	if repo, err = NewRepo(NewBook()); err != nil {
		panic(err)
	}
	if err = repo.CreateRepo(); err != nil {
		panic(err)
	}
	return repo
}

func clearRepo(repo *Repo) {
	if err := repo.DropRepo(); err != nil {
		panic(err)
	}
}
