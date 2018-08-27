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
	CreatedAt time.Time `db:"created_at datetime nil"`
	*Base
}

func (u *User) TableName() string {
	return "user"
}

func NewUser() *User {
	user := NewModel(new(User)).(*User)
	user.DeclareMany("books", new(Book), Nexus{
		"user_id": "id",
		"id":      NWhere{GT, 0},
	})
	return user
}

type Book struct {
	Id     string `db:"id varchar(128) pk"`
	UserId string `db:"user_id varchar(128)"`
	Name   string `db:"name varchar(128)"`
	*Base
}

func (b *Book) TableName() string {
	return "book"
}

func NewBook() *Book {
	book := NewModel(new(Book)).(*Book)
	book.DeclareOne("author", new(User), Nexus{
		"id": "user_id",
	})
	return book
}

type withCustomCount struct {
	data []map[string]interface{}
}

func (wc *withCustomCount) DataOf(m interface{}, _ Nexus) interface{} {
	for _, item := range wc.data {
		if m.(*User).Id == item["user_id"].(string) {
			return item["number"]
		}
	}
	return 0
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

func TestIsModelErr(t *T) {
	if IsModelErr(&Error{ERR_SQL, errors.New("fake error sql")}) {
		log.Printf("%v\t\tOK", "is model err")
	}
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

func TestHas(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		if !user.Has("id") {
			return errors.New("has error when had")
		}
		if user.Has("no") {
			return errors.New("has error when not had")
		}
		return nil
	}, t, "has")
}

func TestGet(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		user.Id = "100"
		if user.Get("id").(string) != "100" {
			return errors.New("get error")
		}
		return nil
	}, t, "get")
}

func TestSet(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		if ok := user.Set("id", "100"); !ok {
			return errors.New("set error")
		}
		if user.Get("id").(string) != "100" {
			return errors.New("set error")
		}
		return nil
	}, t, "set")
}

func TestCreateSlice(t *T) {
	suit(func(t *T) error {
		var err error
		user := NewUser()
		if err = insertUser(user); err != nil {
			return err
		}
		books := []map[string]interface{}{
			{
				"id":           "1",
				"name":         "1",
				"published_at": time.Now(),
				"user_id":      user.Id,
			},
			{
				"id":           "2",
				"name":         "2",
				"published_at": time.Now(),
				"user_id":      user.Id,
			},
		}
		data := []interface{}{}
		for _, b := range books {
			item := NewBook()
			item.Fill(b)
			data = append(data, item)
		}
		return NewBook().Repo().Create(data)

	}, t, "create slice")
}

func TestCreateMap(t *T) {
	suit(func(t *T) error {
		var err error
		user := NewUser()
		if err = insertUser(user); err != nil {
			return err
		}
		books := []map[string]interface{}{
			{
				"id":           "1",
				"name":         "1",
				"published_at": time.Now(),
				"user_id":      user.Id,
			},
			{
				"id":           "2",
				"name":         "2",
				"published_at": time.Now(),
				"user_id":      user.Id,
			},
		}
		data := make(map[interface{}]interface{})
		for _, b := range books {
			item := NewBook()
			item.Fill(b)
			data[b["id"]] = item
		}
		return NewBook().Repo().Create(data)

	}, t, "create map")
}

func TestUpdate(t *T) {
	suit(func(t *T) error {
		var err error
		user := NewUser()
		insertUser(user)
		user.Name = "fixed name"
		if err = user.Save(); err != nil {
			return err
		}
		if u, ok, err := user.Repo().Find("1"); err != nil {
			return err
		} else if ok {
			if u.(*User).Name != "fixed name" {
				return errors.New("save error")
			}
		} else {
			return errors.New("save error")
		}
		return nil
	}, t, "update")
}

func TestCount(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		insertUser(user)
		if count, err := user.Repo().Count(); err != nil {
			return err
		} else if count != 1 {
			return errors.New("count error")
		}
		return nil
	}, t, "count")
}

func TestFetchNexus(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		book := NewBook()
		insertUser(user)
		insertBook(book)
		if users, err := user.Repo().With("books").Fetch(); err != nil {
			return err
		} else {
			for _, user := range users {
				var many interface{}
				if many, err = user.(*User).Many("books"); err != nil {
					return err
				}
				for _, m := range many.(map[interface{}]interface{}) {
					if !isBook(m) {
						return err
					}
				}
			}
		}
		if books, err := book.Repo().With("author").Fetch(); err != nil {
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
		user := NewUser()
		book := NewBook()
		insertUser(user)
		insertBook(book)
		if many, err := user.Many("books"); err != nil {
			return err
		} else {
			for _, m := range many.(map[interface{}]interface{}) {
				if !isBook(m) {
					return err
				}
			}
		}
		return nil
	}, t, "with many")
}

func TestWithCustom(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		book := NewBook()
		insertUser(user)
		insertBook(book)
		user.Repo().WithCustom("books", func(m interface{}) (val NexusValues, err error) {
			repo := m.(Model).Repo()
			repo.Select(E{"count(1) as number"}, "user_id")
			repo.GroupBy("user_id")
			data := []map[string]interface{}{}
			err = repo.Query(func(rows *sql.Rows, _ []string) error {
				var number int
				var user_id string
				if err = rows.Scan(&number, &user_id); err != nil {
					return err
				}
				data = append(data, map[string]interface{}{
					"number":  number,
					"user_id": user_id,
				})
				return nil
			})
			if err == nil {
				val = &withCustomCount{data}
			}
			return
		})
		if ms, err := user.Repo().Fetch(); err != nil {
			return err
		} else {
			for _, m := range ms {
				if count, e := m.(*User).Many("books"); e != nil {
					return e
				} else if count != 1 {
					t.Fatal("with custom count error")
				}
			}
		}
		return nil
	}, t, "with custom")
}

func TestWithOne(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		book := NewBook()
		insertUser(user)
		insertBook(book)
		if one, err := book.One("author"); err != nil {
			return err
		} else if !isUser(one) {
			return errors.New("with one error")
		}
		return nil
	}, t, "with one")
}

func TestFetch(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		insertUser(user)
		if rows, err := user.Repo().Fetch(); err != nil {
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

func TestTx(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		book := NewBook()
		insertUser(user)
		insertBook(book)
		Conn.Tx(func(tx *sql.Tx) error {
			user.Repo().WithTx(tx)
			if err := user.Delete(); err != nil {
				return err
			}
			return errors.New("for test tx")
		}, nil, nil)
		ms := user.Repo().WithoutTx().MustFetch()
		if len(ms) == 0 {
			errors.New("tx error")
		}
		return nil
	}, t, "tx")
}

func TestFind(t *T) {
	suit(func(t *T) error {
		user := NewUser()
		insertUser(user)
		if model, ok, err := user.Repo().Find("1"); err != nil {
			return err
		} else if ok {
			if !isUser(model) {
				return errors.New("find error")
			}
		} else {
			return errors.New("find error")
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

	return user.Create()
}

func insertBook(book *Book) error {
	book.UserId = "1"
	book.Name = "hello world"
	book.Id = "1"

	return book.Create()
}

func createUserRepo() *Repo {
	repo := NewUser().Repo()
	var err error
	if err = repo.CreateRepo(); err != nil {
		panic(err)
	}
	return repo
}

func createBookRepo() *Repo {
	repo := NewBook().Repo()
	var err error
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
		} else if val, err := m.(Mapable).Mapper().colValue(m, fd.colname); err != nil {
			result = false
		} else {
			if val != vals[fd.colname] {
				result = false
			}
		}
		return result
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
	defer func() {
		if e := recover(); e != nil {
			log.Print(e)
		}
	}()
	db := initConn()
	defer db.Close()
	ur := createUserRepo()
	br := createBookRepo()
	err := handle(t)
	clearRepo(ur)
	clearRepo(br)
	if err != nil {
		if IsModelErr(err) {
			log.Print("model error")
		}
		t.Fatalf("%v:\t%v", name, err)
	}
	log.Printf("%v\t\tOK", name)
}
