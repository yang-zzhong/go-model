package model

import (
	"errors"
	"reflect"
)

type NexusValues interface {
	DataOf(m interface{}, rel map[string]string) interface{}
}

type nexusValues struct {
	data map[interface{}]interface{}
}

func (nve *nexusValues) DataOf(m interface{}, rel map[string]string) interface{} {
	result := make(map[interface{}]interface{})
	for k, item := range nve.data {
		eq := true
		for mf, itf := range rel {
			afv, _ := m.(Mapable).Mapper().colValue(m, mf)
			bfv, _ := item.(Mapable).Mapper().colValue(item, itf)
			if afv != bfv {
				eq = false
				break
			}
		}
		if eq {
			result[k] = item
		}
	}
	return result
}

// a mid process needed struct see nexusValues
type fornexus struct {
	m       interface{}
	n       Nexus
	t       int
	handler repoHandler
	where   map[string][]interface{}
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

// append a value to a nexus
func (fn fornexus) append(field string, val interface{}) {
	if _, ok := fn.where[field]; !ok {
		fn.where[field] = []interface{}{}
	}
	fn.where[field] = append(fn.where[field], val)
}

// With tell repo that find nexus defined by model
// if nexus not defined, With will ignore
func (repo *Repo) WithCustom(name string, handler repoHandler) *Repo {
	t := t_bad
	var ok bool
	var m interface{}
	var n map[string]string
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
			data = &nexusValues{d}
		}
		return
	})
}

// nexusValues fetch all nexus result according the repo fetch result
func (repo *Repo) nexusValues(models []interface{}) (result []nexusResult, err error) {
	// find each nexus's query where and model
	mid := make(map[string]fornexus)
	for _, m := range models {
		for _, w := range repo.withs {
			if w.t == t_bad {
				err = errors.New("relationship " + w.name + " not exists")
				return
			}
			if _, ok := mid[w.name]; !ok {
				mid[w.name] = fornexus{
					w.m,
					w.n,
					w.t,
					w.handler,
					make(map[string][]interface{})}
			}
			for af, bf := range w.n {
				val, _ := repo.model.(Mapable).Mapper().colValue(m, af)
				mid[w.name].append(bf, val)
			}
		}
	}
	// fetch nexus result according to mid
	for name, fn := range mid {
		m := fn.m.(Model)
		for field, val := range fn.where {
			m.Repo().WhereIn(field, val)
		}
		if data, e := fn.handler(m); e != nil {
			err = e
			return
		} else {
			result = append(result, nexusResult{name, fn.m, fn.n, fn.t, data})
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
