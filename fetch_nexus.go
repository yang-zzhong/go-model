package model

import (
	"errors"
	"reflect"
)

type NWhere struct {
	Op    string
	Value interface{}
}

type NexusValues interface {
	DataOf(m interface{}, rel Nexus) interface{}
}

type DefaultNexusValues struct {
	data map[interface{}]interface{}
}

func NewNexusValues(data map[interface{}]interface{}) NexusValues {
	return &DefaultNexusValues{data}
}

func (nve *DefaultNexusValues) DataOf(m interface{}, rel Nexus) interface{} {
	result := make(map[interface{}]interface{})
	for k, item := range nve.data {
		eq := true
		for field, val := range rel {
			switch val.(type) {
			case string:
				a, _ := m.(Mapable).Mapper().colValue(m, val.(string))
				b, _ := item.(Mapable).Mapper().colValue(item, field)
				if a != b {
					eq = false
					break
				}
			}
		}
		if eq {
			result[k] = item
		}
	}
	return result
}

// a nexus result struct to hold the query result of a nexus
type nexusResult struct {
	name string
	m    interface{}
	n    Nexus
	t    int
	data NexusValues
}

type repoHandler func(m interface{}) (NexusValues, error)
type fornexusHandler func(field, op string, value interface{})

// With tell repo that find nexus defined by model
// if nexus not defined, With will ignore
func (repo *Repo) WithCustom(name string, handler repoHandler) *Repo {
	t := t_bad
	var ok bool
	var m interface{}
	var n map[string]interface{}
	if m, n, ok = repo.model.(NexusOne).HasOne(name); ok {
		t = t_one
	} else if m, n, ok = repo.model.(NexusMany).HasMany(name); ok {
		t = t_many
	}
	repo.withs = append(repo.withs, with{
		name:    name,
		m:       m,
		n:       n,
		t:       t,
		handler: handler,
	})
	return repo
}

func (repo *Repo) With(name string) *Repo {
	return repo.WithCustom(name, func(m interface{}) (data NexusValues, err error) {
		if d, e := m.(Model).Repo().FetchKey(m.(Model).PK()); e != nil {
			err = e
		} else {
			data = &DefaultNexusValues{d}
		}
		return
	})
}

// nexusValues fetch all nexus result according the repo fetch result
func (repo *Repo) nexusValues(models []interface{}) (result []nexusResult, err error) {
	// find each nexus's query where and model
	if len(models) == 0 {
		return
	}
	for _, w := range repo.withs {
		if w.t == t_bad {
			err = errors.New("relationship " + w.name + " not exists")
			return
		}
		r := w.m.(Model).Repo()
		for af, bf := range w.n {
			switch bf.(type) {
			case NWhere:
				r.Where(af, bf.(NWhere).Op, bf.(NWhere).Value)
			case string:
				vals := []interface{}{}
				for _, m := range models {
					if val, e := repo.model.(Mapable).Mapper().colValue(m, bf.(string)); e != nil {
						err = e
						return
					} else {
						vals = append(vals, val)
					}
				}
				if len(vals) != 0 {
					r.WhereIn(af, vals)
				}
			}
		}
		if data, e := w.handler(w.m); e != nil {
			err = e
			return
		} else {
			result = append(result, nexusResult{w.name, w.m, w.n, w.t, data})
		}
	}

	return
}

//
// bind nexus result to each fetched model
//
func (repo *Repo) bindNexus(m interface{}, nr []nexusResult) {
	for _, n := range nr {
		nm := n.data.DataOf(m, n.n)
		if n.t == t_many {
			m.(NexusMany).SetMany(n.name, nm)
			continue
		}
		ty := reflect.TypeOf(nm)
		switch ty.Kind() {
		case reflect.Map:
			val := nm.(map[interface{}]interface{})
			for _, item := range val {
				m.(NexusOne).SetOne(n.name, item)
				break
			}
		case reflect.Slice:
			val := nm.([]interface{})
			for _, item := range val {
				m.(NexusOne).SetOne(n.name, item)
				break
			}
		default:
			m.(NexusOne).SetOne(n.name, nm)
		}
	}
}
