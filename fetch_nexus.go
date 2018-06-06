package model

import (
	"errors"
)

// a mid process needed struct see nexusValues
type fornexus struct {
	m     interface{}
	n     Nexus
	t     int
	where map[string][]interface{}
}

// a nexus result struct to hold the query result of a nexus
type nexusResult struct {
	name string
	m    interface{}
	n    Nexus
	t    int
	data map[interface{}]interface{}
}

// append a value to a nexus
func (fn fornexus) append(field string, val interface{}) {
	if _, ok := fn.where[field]; !ok {
		fn.where[field] = []interface{}{}
	}
	fn.where[field] = append(fn.where[field], val)
}

// With tell repo that find nexus defined by model
// if nexus not defined, With will ignore
func (repo *Repo) With(name string) *Repo {
	if m, n, ok := repo.model.(NexusOne).HasOne(name); ok {
		repo.withs = append(repo.withs, with{name, m, n, t_one})
	} else if m, n, ok := repo.model.(NexusMany).HasMany(name); ok {
		repo.withs = append(repo.withs, with{name, m, n, t_many})
	} else {
		repo.withs = append(repo.withs, with{name, nil, make(map[string]string), t_bad})
	}
	return repo
}

// nexusValues fetch all nexus result according the repo fetch result
func (repo *Repo) nexusValues(models map[interface{}]interface{}) (result []nexusResult, err error) {
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
					make(map[string][]interface{})}
			}
			for af, bf := range w.n {
				val, _ := repo.model.(Mapable).Mapper().ColValue(m, af)
				mid[w.name].append(bf, val)
			}
		}
	}
	// fetch nexus result according to mid
	for name, fn := range mid {
		r, _ := NewRepo(fn.m)
		for field, val := range fn.where {
			r.WhereIn(field, val)
		}
		if data, perr := r.Fetch(); perr == nil {
			result = append(result, nexusResult{name, fn.m, fn.n, fn.t, data})
		} else {
			err = perr
			return
		}
	}

	return
}

//
// bind nexus result to each fetched model
//
func (repo *Repo) bindNexus(m interface{}, nr []nexusResult) {
	manys := make(map[string]map[interface{}]interface{})
	for _, n := range nr {
		for id, nm := range n.data {
			eq := true
			for af, bf := range n.n {
				afv, _ := repo.model.(Mapable).Mapper().ColValue(m, af)
				bfv, _ := n.m.(Mapable).Mapper().ColValue(nm, bf)
				if afv != bfv {
					eq = false
					break
				}
			}
			if !eq {
				continue
			}
			if n.t == t_one {
				m.(NexusOne).SetOne(n.name, nm)
				continue
			}
			if _, ok := manys[n.name]; !ok {
				manys[n.name] = make(map[interface{}]interface{})
			}
			manys[n.name][id] = nm
		}
	}
	for name, data := range manys {
		m.(NexusMany).SetMany(name, data)
	}
}
