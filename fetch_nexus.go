package model

type fornexus struct {
	m     interface{}
	n     Nexus
	t     int
	where map[string][]interface{}
}

type nexusResult struct {
	name string
	n    Nexus
	t    int
	data map[interface{}]interface{}
}

func (fn fornexus) append(field string, val interface{}) {
	if _, ok := fn.where[field]; !ok {
		fn.where[field] = []interface{}{}
	}
	fn.where[field] = append(fn.where[field], val)
}

func (repo *Repo) WithOne(name string) *Repo {
	if m, n, ok := repo.model.(NexusOne).HasOne(name); ok {
		repo.withs = append(repo.withs, with{name, m, n, t_one})
	}
	return repo
}

func (repo *Repo) WithMany(name string) *Repo {
	if m, n, ok := repo.model.(NexusMany).HasMany(name); ok {
		repo.withs = append(repo.withs, with{name, m, n, t_many})
	}
	return repo
}

func (repo *Repo) nexusValues(models map[interface{}]interface{}) []nexusResult {
	mid := make(map[string]fornexus)
	for _, m := range models {
		for _, w := range repo.withs {
			if _, ok := mid[w.name]; !ok {
				mid[w.name] = fornexus{
					w.m,
					w.n,
					w.t,
					make(map[string][]interface{})}
			}
			for af, bf := range w.n {
				val, _ := repo.mm.ColValue(m, af)
				mid[w.name].append(bf, val)
			}
		}
	}
	result := []nexusResult{}
	for name, fn := range mid {
		r, _ := NewRepo(fn.m)
		for field, val := range fn.where {
			r.WhereIn(field, val)
		}
		if data, err := r.Fetch(); err == nil {
			result = append(result, nexusResult{name, fn.n, fn.t, data})
		}
	}

	return result
}

func (repo *Repo) bindNexus(m interface{}, nr []nexusResult) {
	manys := make(map[string]map[interface{}]interface{})
	for _, n := range nr {
		for id, nm := range n.data {
			nmm := NewModelMapper(nm)
			eq := true
			for af, bf := range n.n {
				afv, _ := repo.mm.ColValue(m, af)
				bfv, _ := nmm.ColValue(nm, bf)
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
