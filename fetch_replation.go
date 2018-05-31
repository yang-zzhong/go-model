package model

func (repo *Repo) columnValues() map[string][]interface{} {
	columnValues := make(map[string][]interface{})
	for _, with := range repo.with {
		var rel map[string]string
		if with.t == ONE {
			if _, rel, err = repo.m.(Model).WithOne(name); err != nil {
				return
			}
		} else {
			if _, rel, err = repo.m.(Model).WithMany(name); err != nil {
				return
			}
		}
		for field, _ := range rel {
			conlumnValues[field] = []interface{}{}
		}
	}

	return columnValues
}

func (repo *Repo) voc(m interface{}, cv map[string][]interface{}) error {
	for field, _ := range cv {
		var val interface{}
		if val, err = repo.mm.FieldValue(m, field); err != nil {
			return
		}
		cv[field] = append(cv[field], val)
	}
	return
}

func (repo *Repo) attachModel(m Model) map[string]interface{} {
}

func (repo *Repo) ones(columnValues map[string][]interface{}) (result []map[string]interface{}, err error) {
	for _, name := range repo.withOne {
		var one interface{}
		var rel map[string]string
		if one, rel, err = repo.m.(Model).WithOne(name); err != nil {
			return
		}
		repo := NewCustomRepo(one, repo.conn, repo.modifier)
		for a, b := range rel {
			repo.WhereIn(b, columnValues[a])
		}
		m := repo.One()
		item := make(map[string]interface{})
	}
}
