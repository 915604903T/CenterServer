package handlers

type UnionSet struct {
	Parent map[string]string
	Rank   map[string]int
}

func NewUnionSet() UnionSet {
	return UnionSet{
		Parent: make(map[string]string),
		Rank:   make(map[string]int),
	}
}

func (u *UnionSet) initInsert(x string) bool {
	if _, ok := u.Parent[x]; ok {
		return false
	}
	u.Parent[x] = x
	u.Rank[x] = 0
	return true
}

func (u *UnionSet) find(x string) string {
	if u.Parent[x] != x {
		u.Parent[x] = u.find(u.Parent[x])
	}
	return u.Parent[x]
}

func (u *UnionSet) union(x, y string) bool {
	px, py := u.find(x), u.find(y)
	if px == py {
		return false
	}
	if u.Rank[px] < u.Rank[py] {
		u.Parent[px] = py
	} else {
		u.Parent[py] = px
		if u.Rank[py] == u.Rank[px] {
			u.Rank[px]++
		}
	}
	return true
}
