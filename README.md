## go model and repo

### sample
```go

import (
    . "github.com/yang-zzhong/go-querybuilder"
    . "github.com/yang-zzhong/go-model"
    helpers "github.com/yang-zzhong/go-helpers"
    "database/sql"
)

// define a model
type User struct {
    Id          string      `db:"id char(36) pk"`
    Account     string      `db:"account varchar(36) uk"`
    Name        string      `db:"name varchar(36)"`
    Address     string      `db:"address varchar(128) nil"`
    Birthday    time.Time   `db:"birthday datetime"`
}

func (user *User) PK() string {
    return "id"
}

func (user *User) TableName() string {
    return "user"
}

// 获取db
func driver() *sql.DB {
    drv, err := sql.Open("mysql", "mysql_user:password@/db?parseTime=true")
    if err != nil {
        panic(err)
    }

    return drv
}

// model's construct func
func NewUser() {
    user := new(User)
    user.Id = helpers.RandString(32)
    return user
}

// repo's construct func
func NewUserRepo() (repo *Repo, err error) {
    repo, err = NewRepo(new(User), driver(), &MysqlModifier)
}


repo, err := NewUserRepo()
if err != nil {
    panic(err)
}

// 创建表
repo.CreateTable()

model := NewUser()
model.Account = "my-account-name"
model.Name = "my-name"
model.Birthday = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

// 创建model
if err := repo.Create(model); err != nil {
    panic(err)
}

item := repo.Find("id value")
if item == nil {
    panic("user not found")
}
user := item.(*User)

user.Name = "another name"

// 更新model
if err := repo.Update(user); err != nil {
    panic(err)
}

// 根据条件找model

repo.Where("name", LIKE, "name").Quote(func(repo *Repo) {
    repo.Where("birthday", GT, time.Date(2000, time.November, 0, 0, 0, 0, 0, time.UTC))
    repo.Or().Where("birthday", LT, time.Date(1990, time.November, 0, 0, 0, 0, 0, time.UTC))
})

repo.Count()        // count
// repo.One()       // 取第一个
// repo.Fetch()     // 取所有数据

```
