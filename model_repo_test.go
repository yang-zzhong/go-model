package model

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/yang-zzhong/go-querybuilder"
	"log"
	. "testing"
	"time"
)

type handler func(*T) error

type User struct {
	Id        string    `db:"id varchar(128) pk"`
	Name      string    `db:"name varchar(32) uk"`
	Age       int       `db:"age int"`
	Level     int       `db:"level int"`
	Optional  string    `db:"optional varchar(256) nil"`
	CreatedAt time.Time `db:"created_at datetime"`
	*Base
}

type Book struct {
	Id     string `db:"id varchar(128) pk"`
	UserId string `db:"user_id varchar(128)"`
	Name   string `db:"name varchar(128)"`
	*Base
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
	return One(u.Base, u, name)
}

func (u *User) Many(name string) (map[interface{}]interface{}, error) {
	return Many(u.Base, u, name)
}

func (b *Book) Many(name string) (map[interface{}]interface{}, error) {
	return Many(b.Base, b, name)
}

func (b *Book) One(name string) (interface{}, error) {
	return One(b.Base, b, name)
}

func NewUser() *User {
	user := new(User)
	user.Base = NewBase(user)
	book := new(Book)
	book.Base = NewBase(book)
	user.DeclareMany("books", book, map[string]string{
		"id": "user_id",
	})
	return user
}

func NewBook() *Book {
	book := new(Book)
	book.Base = NewBase(book)
	user := new(User)
	user.Base = NewBase(user)
	book.DeclareOne("author", user, map[string]string{
		"user_id": "id",
	})
	return book
}

func TestFetchNexus(t *T) {
	suit(func(t *T) error {
		var users map[interface{}]interface{}
		var books map[interface{}]interface{}
		var err error
		var ur, br *Repo
		if ur, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if br, err = NewRepo(NewBook()); err != nil {
			return err
		}
		if _, err = insertUser(ur); err != nil {
			return err
		}
		if _, err := insertBook(br); err != nil {
			return err
		}
		if users, err = ur.WithMany("books").Fetch(); err != nil {
			return err
		}
		for _, user := range users {
			u := user.(User)
			var many map[interface{}]interface{}
			if many, err = (&u).Many("books"); err != nil {
				return err
			}
			for _, m := range many {
				if !isBook(m) {
					return err
				}
			}
		}
		if books, err = br.WithOne("author").Fetch(); err != nil {
			return err
		}
		for _, book := range books {
			b := book.(Book)
			var one interface{}
			if one, err = (&b).One("author"); err != nil {
				return err
			}
			if !isUser(one) {
				return err
			}
		}
		return nil
	}, t, "fetch nexus")
}

func TestWithMany(t *T) {
	suit(func(t *T) error {
		var user *User
		var err error
		var ur, br *Repo
		if ur, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if br, err = NewRepo(NewBook()); err != nil {
			return err
		}
		if user, err = insertUser(ur); err != nil {
			return err
		}
		if _, err := insertBook(br); err != nil {
			return err
		}
		var many map[interface{}]interface{}
		if many, err = user.Many("books"); err != nil {
			return err
		}
		for _, m := range many {
			if !isBook(m) {
				return err
			}
		}
		return nil
	}, t, "with many")
}

func TestWithOne(t *T) {
	suit(func(t *T) error {
		var book *Book
		var err error
		var ur, br *Repo
		if ur, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if br, err = NewRepo(NewBook()); err != nil {
			return err
		}
		if _, err = insertUser(ur); err != nil {
			return err
		}
		if book, err = insertBook(br); err != nil {
			return err
		}
		var one interface{}
		if one, err = book.One("author"); err != nil {
			return err
		}
		if !isUser(one) {
			return errors.New("with one error")
		}
		return nil
	}, t, "with one")
}

func TestCreate(t *T) {
	suit(func(t *T) error {
		var repo *Repo
		var err error
		if repo, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if _, err := insertUser(repo); err != nil {
			return err
		}
		return nil
	}, t, "create")
}

func TestFetch(t *T) {
	suit(func(t *T) error {
		var repo *Repo
		var err error
		if repo, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if _, err := insertUser(repo); err != nil {
			return err
		}
		if rows, err := repo.Fetch(); err != nil {
			return err
		} else {
			for _, row := range rows {
				if !isUser(row) {
					return errors.New("fetch error")
				}
			}
		}
		return nil
	}, t, "fetch")
}

func TestFind(t *T) {
	suit(func(t *T) error {
		var repo *Repo
		var err error
		if repo, err = NewRepo(NewUser()); err != nil {
			return err
		}
		if _, err := insertUser(repo); err != nil {
			return err
		}
		if model, err := repo.Find("1"); err != nil {
			return err
		} else {
			if !isUser(model) {
				return errors.New("find error")
			}
		}
		return nil
	}, t, "find")
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

func isUser(m interface{}) bool {
	if _, ok := m.(User); !ok {
		return false
	}
	vals := map[string]interface{}{
		"id":    "1",
		"name":  "yang-zhong",
		"age":   17,
		"level": 1,
	}

	return rightModel(m, vals)
}

func isBook(m interface{}) bool {
	if _, ok := m.(Book); !ok {
		return false
	}
	vals := map[string]interface{}{
		"id":   "1",
		"name": "hello world",
	}

	return rightModel(m, vals)
}

func rightModel(m interface{}, vals map[string]interface{}) bool {
	result := true
	m.(Mapable).Mapper().each(func(fd *fieldDescriptor) bool {
		if _, ok := vals[fd.colname]; !ok {
			return true
		}
		if val, err := m.(Mapable).Mapper().ColValue(m, fd.colname); err != nil {
			result = false
			return false
		} else {
			if val != vals[fd.colname] {
				result = false
				return false
			}
		}
		return true
	})

	return result
}

func initConn() *sql.DB {
	var err error
	var con *sql.DB
	con, err = sql.Open("mysql", "root:young159357789@/test_go?parseTime=true")
	if err != nil {
		panic(err)
	}
	Config(con, &MysqlModifier{})

	return con
}

func suit(handle handler, t *T, name string) {
	db := initConn()
	defer db.Close()
	ur := createUserRepo()
	br := createBookRepo()
	log.Print("begin: " + name)
	err := handle(t)
	log.Print("end: " + name)
	clearRepo(ur)
	clearRepo(br)
	if err != nil {
		t.Fatal(err)
	}
}