package jira

type ProjectsMeta struct {
	Projects map[string]ProjectMeta
}

func NewProjectsMeta() ProjectsMeta {
	pm := ProjectsMeta{}
	pm.init()
	return pm
}

func (pm *ProjectsMeta) init() {
	if pm.Projects == nil {
		pm.Projects = map[string]ProjectMeta{}
	}
}

func (pm *ProjectsMeta) AddMap(info map[string]float32) {
	pm.init()
	for key, teamSize := range info {
		proj, ok := pm.Projects[key]
		if !ok {
			proj = ProjectMeta{Key: key}
		}
		proj.TeamSize = teamSize
		pm.Projects[key] = proj
	}
}

func (pm *ProjectsMeta) FTEs() float32 {
	sum := float32(0)
	for _, projMeta := range pm.Projects {
		sum += projMeta.TeamSize
	}
	return sum
}

func (pm *ProjectsMeta) CapacitySimple(itemsPerWeekPerFTE, weekCount float32) float32 {
	ftes := pm.FTEs()
	return ftes * itemsPerWeekPerFTE * weekCount
}

type ProjectMeta struct {
	Key      string
	TeamSize float32
}
