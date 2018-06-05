package model

import (
	"database/sql"
	"encoding/json"
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

func (b *Book) TableName() string {
	return "book"
}

func NewUser() *User {
	user := NewModel(new(User)).(*User)
	user.DeclareMany("books", new(Book), map[string]string{
		"id": "user_id",
	})
	return user
}

func NewBook() *Book {
	book := NewModel(new(Book)).(*Book)
	book.DeclareOne("author", new(User), map[string]string{
		"user_id": "id",
	})
	return book
}

func TestFill(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		user.Fill(map[string]interface{}{
			"id":        "1",
			"name":      "yang-zhong",
			"age":       17,
			"level":     1,
			"create_at": time.Now(),
		})
		if !isUser(user) {
			return errors.New("fill error")
		}
		return nil
	}, t, "fill")
}

func TestFetchNexus(t *T) {
	suit(func(t *T) error {
		var err error
		var ur, br *Repo
		user := NewUser()
		book := NewBook()
		if err = insertUser(user); err != nil {
			return err
		}
		if err := insertBook(book); err != nil {
			return err
		}
		if ur, err = user.Repo(); err != nil {
			return err
		}
		if br, err = book.Repo(); err != nil {
			return err
		}
		if users, err := ur.With("books").Fetch(); err != nil {
			return err
		} else {
			for _, user := range users {
				var many map[interface{}]interface{}
				if many, err = user.(*User).Many("books"); err != nil {
					return err
				}
				for _, m := range many {
					if !isBook(m) {
						return err
					}
				}
			}
		}
		if books, err := br.With("author").Fetch(); err != nil {
			return err
		} else {
			for _, book := range books {
				var one interface{}
				if one, err = book.(*Book).One("author"); err != nil {
					return err
				}
				if !isUser(one) {
					return err
				}
			}
		}
		return nil
	}, t, "fetch nexus")
}

func TestWithMany(t *T) {
	suit(func(t *T) error {
		var err error
		user := NewUser()
		book := NewBook()
		if err = insertUser(user); err != nil {
			return err
		}
		if err := insertBook(book); err != nil {
			return err
		}
		if many, err := user.Many("books"); err != nil {
			return err
		} else {
			for _, m := range many {
				if !isBook(m) {
					return err
				}
			}
		}
		return nil
	}, t, "with many")
}

func TestWithOne(t *T) {
	suit(func(t *T) error {
		var err error
		user := NewUser()
		book := NewBook()
		if err = insertUser(user); err != nil {
			return err
		}
		if err = insertBook(book); err != nil {
			return err
		}
		if one, err := book.One("author"); err != nil {
			return err
		} else if !isUser(one) {
			return errors.New("with one error")
		}
		return nil
	}, t, "with one")
}

func TestCreate(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		if err := insertUser(user); err != nil {
			return err
		}
		return nil
	}, t, "create")
}

func TestFetch(t *T) {
	suit(func(t *T) error {
		var repo *Repo
		var err error
		user := NewUser()
		if repo, err = user.Repo(); err != nil {
			return err
		}
		if err := insertUser(user); err != nil {
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
		user := NewUser()
		if repo, err = user.Repo(); err != nil {
			return err
		}
		if err := insertUser(user); err != nil {
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

func TestMarsha1(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		insertUser(user)
		if _, err := json.Marshal(user); err != nil {
			return err
		}
		return nil
	}, t, "marsha1")
}

func insertUser(user *User) error {
	user.Id = "1"
	user.Name = "yang-zhong"
	user.Age = 17
	user.Level = 1
	user.CreatedAt = time.Now()

	return user.Create()
}

func insertBook(book *Book) error {
	book.UserId = "1"
	book.Name = "hello world"
	book.Id = "1"

	return book.Create()
}

func createUserRepo() *Repo {
	var repo *Repo
	var err error
	if repo, err = NewUser().Repo(); err != nil {
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
	if repo, err = NewBook().Repo(); err != nil {
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
	if _, ok := m.(*User); !ok {
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
	if _, ok := m.(*Book); !ok {
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
	err := handle(t)
	clearRepo(ur)
	clearRepo(br)
	if err != nil {
		t.Fatal(err)
	}
	log.Print(name + ": OK ^_^")
}
