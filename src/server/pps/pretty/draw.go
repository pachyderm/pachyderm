package pretty

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type renderConfig struct {
	boxWidth   int
	edgeHeight int
}

type RenderOption func(*renderConfig)

func BoxWidthOption(boxWidth int) RenderOption {
	return func(ro *renderConfig) {
		ro.boxWidth = boxWidth
	}
}

func EdgeHeightOption(verticalSpace int) RenderOption {
	return func(ro *renderConfig) {
		ro.edgeHeight = verticalSpace
	}
}

type vertex struct {
	id        string
	label     string
	edges     map[string]*vertex
	layer     int
	rowOffset int
	red       bool
}

func newVertex(label string) *vertex {
	return &vertex{id: uuid.New().String(), label: label, edges: make(map[string]*vertex)}
}

func dummyVertex() *vertex {
	return &vertex{id: uuid.New().String(), label: "*", edges: make(map[string]*vertex)}
}

func (v *vertex) addEdge(u *vertex) {
	v.edges[u.id] = u
}

func (v *vertex) removeEdge(u *vertex) {
	delete(v.edges, u.id)
}

type layerer func([]*vertex) [][]*vertex
type orderer func([][]*vertex)

func Draw(pis []*pps.PipelineInfo, opts ...RenderOption) (string, error) {
	ro := &renderConfig{
		boxWidth:   11,
		edgeHeight: 5,
	}
	for _, o := range opts {
		o(ro)
	}
	g, err := makeGraph(pis)
	if err != nil {
		return "", err
	}
	return draw(g, layerLongestPath, orderGreedy, ro), nil
}

func makeGraph(pis []*pps.PipelineInfo) ([]*vertex, error) {
	vMap := make(map[string]*vertex)
	vs := make([]*vertex, 0)
	upsertVertex := func(name string, lastState pps.JobState) *vertex {
		v := newVertex(name)
		if _, ok := vMap[name]; !ok {
			vMap[name] = v
			vs = append(vs, v)
		}
		if lastState == pps.JobState_JOB_FAILURE || lastState == pps.JobState_JOB_KILLED {
			vMap[name].red = true
		}
		return vMap[name]
	}
	for _, pi := range pis {
		pv := upsertVertex(pi.Pipeline.Name, pi.LastJobState)
		if err := pps.VisitInput(pi.Details.Input, func(input *pps.Input) error {
			var name string
			if input.Pfs != nil {
				name = input.Pfs.Name
			} else if input.Cron != nil {
				name = input.Cron.Name
			} else {
				return nil
			}
			iv := upsertVertex(name, pps.JobState_JOB_STATE_UNKNOWN)
			iv.addEdge(pv)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return vs, nil
}

func draw(vertices []*vertex, lf layerer, of orderer, ro *renderConfig) string {
	// Assign Layers
	layers := lf(vertices)
	of(layers)
	assignCoordinates(layers, ro)
	picture := renderPicture(layers, ro)
	return picture
}

// precompute the box coordinates so that during rendering the edges can be filled between layers
func assignCoordinates(layers [][]*vertex, ro *renderConfig) {
	maxWidth := rowWidth(layers, ro)
	for _, l := range layers {
		boxCenterOffset := maxWidth / (len(l) + 1)
		for j := 0; j < len(l); j++ {
			l[j].rowOffset = (j + 1) * boxCenterOffset
		}
	}
}

// ==================================================
// Layering Algorithms

func layerLongestPath(vs []*vertex) [][]*vertex {
	assigned := make(map[string]*vertex, 0)
	var layers [][]*vertex
	addToLayer := func(v *vertex, l int) {
		if l >= len(layers) {
			layers = append(layers, make([]*vertex, 0))
		}
		layers[l] = append(layers[l], v)
	}
	fillDummies := func(v *vertex) {
		// we gather list of callbacks so that we don't mutate v.edges as we are iterating over it
		cbs := make([]func(), 0)
		for _, e := range v.edges {
			diff := v.layer - e.layer
			if diff > 1 {
				u := e // necessary to copy the basic vertex fields
				cbs = append(cbs, func() {
					latest := v
					for i := 0; i < diff-1; i++ {
						d := dummyVertex()
						d.addEdge(u)
						addToLayer(d, v.layer-i-1)
						latest.removeEdge(u)
						latest.addEdge(d)
						latest = d
					}
				})
			}
		}
		for _, cb := range cbs {
			cb()
		}
	}
	// build the layers up from the bottom of the DAG
	for _, v := range leaves(vs) {
		assigned[v.id] = v
		addToLayer(v, 0)
	}
	for len(vs) != len(assigned) {
		for _, v := range vs {
			func() {
				if _, ok := assigned[v.id]; !ok {
					var maxLevel int
					// check this node is assignable
					for _, e := range v.edges {
						u, eDone := assigned[e.id]
						if !eDone {
							return
						}
						maxLevel = max(u.layer, maxLevel)
					}
					v.layer = maxLevel + 1
					addToLayer(v, v.layer)
					assigned[v.id] = v
					fillDummies(v)
				}
			}()
		}
	}
	return layers
}

// ==================================================
// Ordering Algorithms

func orderGreedy(layers [][]*vertex) {
	var prev map[string]int
	for i, l := range layers {
		if i > 0 {
			sort.Slice(l, func(i, j int) bool {
				iScore, jScore := 0, 0
				for u := range l[i].edges {
					iScore += prev[u] - i
				}
				for u := range l[j].edges {
					jScore += prev[u] - j
				}
				return iScore < jScore
			})
		}
		prev = make(map[string]int)
		for j, v := range l {
			prev[v.id] = j
		}
	}
}

// ==================================================
// Rendering algorithm

func renderPicture(layers [][]*vertex, ro *renderConfig) string {
	picture := ""
	maxRowWidth := rowWidth(layers, ro)
	// traverse the layers starting with source repos
	for i := len(layers) - 1; i >= 0; i-- {
		l := layers[i]
		written := 0
		row, border := "", ""
		renderEdges := make([]renderEdge, 0)
		// print the row of boxed vertices
		for j := 0; j < len(l); j++ {
			v := l[j]
			colorSprint := color.New(color.FgHiGreen).SprintFunc()
			if v.red {
				colorSprint = color.New(color.FgHiRed).SprintFunc()
			}
			spacing := v.rowOffset - (ro.boxWidth+2)/2 - written
			label := v.label
			if len(label) > ro.boxWidth {
				label = label[:ro.boxWidth-2] + ".."
			}
			boxPadLeft := strings.Repeat(" ", (ro.boxWidth-len(label))/2)
			boxPadRight := strings.Repeat(" ", ro.boxWidth-len(label)-len(boxPadLeft))
			if v.label == "*" {
				hiddenRow := fmt.Sprintf("%s %s%s%s ", strings.Repeat(" ", spacing), boxPadLeft, "|", boxPadRight)
				border += hiddenRow
				row += hiddenRow
			} else {
				border += colorSprint(fmt.Sprintf("%s+%s+", strings.Repeat(" ", spacing), strings.Repeat("-", ro.boxWidth)))
				row += colorSprint(fmt.Sprintf("%s|%s%s%s|", strings.Repeat(" ", spacing), boxPadLeft, label, boxPadRight))
			}
			written += spacing + len(boxPadLeft) + len(label) + len(boxPadRight) + 2
			for _, u := range v.edges {
				renderEdges = append(renderEdges, renderEdge{src: v.rowOffset, dest: u.rowOffset})
			}
		}
		picture += fmt.Sprintf("%s\n%s\n%s\n", border, row, border)
		// print up to `layerVerticalSpace` rows that will contain edge drawings
		for j := 0; j < ro.edgeHeight; j++ {
			row := strings.Repeat(" ", maxRowWidth)
			for _, re := range renderEdges {
				row = re.render(row, j, ro.edgeHeight)
			}
			picture += fmt.Sprint(row)
			picture += "\n"
		}
	}
	return picture
}

// renderEdge is used to describe the source and destination of an edge in terms of the x-axis.
// The number of vertical lines spanned is calculated at each layer
type renderEdge struct {
	src  int
	dest int
}

func (re renderEdge) distance() int {
	return abs(re.src - re.dest)
}

// render sets an edge character in the 'row' string, calculated using the re.src & re.dest (the edge's range along the x-axis),
// and 'vertDist' (the height of the edge) and 'vertIdx' (how many lines down 'vertDist' we are)
func (re renderEdge) render(row string, vertIdx, vertDist int) string {
	setEdgeChar := func(s string, i int, r rune) string {
		if s[i] == byte(r) {
			return s
		} else if s[i] == ' ' {
			return s[:i] + string(r) + s[i+1:]
		}
		return s[:i] + "+" + s[i+1:] // set the coordinate to "+" if there's an edge crossing
	}
	if re.src == re.dest {
		return setEdgeChar(row, re.src, '|')
	}
	const srcEdgeCenterOffset = 1                        // start drawing a diagonal edge one space away from the center of a node
	adjustedXDist := re.distance() - srcEdgeCenterOffset // number of horizontal spaces we must fill with edges
	// vertical line
	if vertDist > adjustedXDist && vertIdx >= adjustedXDist/2 && vertIdx < vertDist-adjustedXDist/2 {
		return setEdgeChar(row, (re.src+re.dest)/2, '|')
	}
	// horizontal line
	if vertDist < adjustedXDist && vertIdx == vertDist/2 {
		step := 1
		if re.src > re.dest {
			step = -1
		}
		diagCoverage := (vertDist / 2) * step
		tmp := re.src + diagCoverage
		for tmp != re.dest-diagCoverage-(step*vertDist%2) {
			tmp += step
			row = setEdgeChar(row, tmp, '-')
		}
		return row
	}
	// diagonal line
	offset := vertIdx + srcEdgeCenterOffset
	// calculate offset based on distance from the end in case we're on the bottom half of the edge
	if vertIdx > vertDist/2 {
		offset = adjustedXDist - (vertDist - offset)
	}
	if re.src > re.dest {
		return setEdgeChar(row, re.src-offset, '/')
	} else {
		return setEdgeChar(row, re.src+offset, '\\')
	}
}

// ==================================================

func leaves(vs []*vertex) []*vertex {
	ls := make([]*vertex, 0)
	for _, v := range vs {
		if len(v.edges) == 0 {
			ls = append(ls, v)
		}
	}
	return ls
}

func rowWidth(layers [][]*vertex, ro *renderConfig) int {
	mlw := maxLayerWidth(layers)
	return mlw * (ro.boxWidth + 2) * 2
}

func maxLayerWidth(layers [][]*vertex) int {
	m := 0
	for _, l := range layers {
		m = max(m, len(l))
	}
	return m
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return x * -1
	}
	return x
}
