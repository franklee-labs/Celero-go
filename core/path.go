package core

import "sort"

type PathGroup struct {
	paths []*Path
}

func NewPathGroup(paths ...*Path) *PathGroup {
	if len(paths) == 0 {
		return &PathGroup{paths: make([]*Path, 0)}
	}
	return &PathGroup{paths: paths}
}

func (g *PathGroup) AppendPath(path *Path) {
	if g.paths == nil {
		g.paths = make([]*Path, 0)
	}
	g.paths = append(g.paths, path)
}

func (g *PathGroup) Paths() []*Path {
	return g.paths
}

func (g *PathGroup) Size() int {
	return len(g.paths)
}

func (g *PathGroup) Get(idx int) *Path {
	if idx < 0 || idx >= g.Size() {
		return nil
	}
	return g.paths[idx]
}

func (g *PathGroup) Sort() {
	if len(g.paths) == 0 {
		return
	}
	for _, p := range g.paths {
		p.Sort()
	}
}

type Path struct {
	conditions []Condition
}

func NewPath(conditions []Condition) *Path {
	return &Path{conditions: conditions}
}

func (p *Path) Conditions() []Condition {
	return p.conditions
}

func (p *Path) Size() int {
	return len(p.conditions)
}

func (p *Path) Get(idx int) Condition {
	if idx < 0 || idx >= p.Size() {
		return nil
	}
	return p.conditions[idx]
}

func (p *Path) Sort() {
	if len(p.conditions) == 0 {
		return
	}
	sort.Slice(p.conditions, func(i, j int) bool {
		return p.conditions[i].Priority() < p.conditions[j].Priority()
	})
}
