package pretty

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// maxLabelLen        = 10
	boxWidth           = 11
	padding            = 10
	layerVerticalSpace = 5
)

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

func (v *vertex) String() string {
	return v.label
}

type layerer func([]*vertex) [][]*vertex
type orderer func([][]*vertex)

func Draw(pis []*pps.PipelineInfo) (string, error) {
	if g, err := makeGraph(pis); err != nil {
		return "", err
	} else {
		return draw(g, layerLongestPath, simpleOrder), nil
	}
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
		if lastState == pps.JobState_JOB_FAILURE {
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

func draw(vertices []*vertex, lf layerer, of orderer) string {
	// Assign Layers
	layers := lf(vertices)
	of(layers)
	assignCoordinates(layers)
	picture := renderPicture(layers)
	return picture
}

// precompute the box coordinates so that during rendering the edges can be filled between layers
func assignCoordinates(layers [][]*vertex) {
	maxWidth := rowWidth(layers)
	for i := 0; i < len(layers); i++ {
		l := layers[i]
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
					// check this node isassignable
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

// TODO: write ordering algorithm
func simpleOrder(layers [][]*vertex) {
}

// ==================================================
// Rendering algorithm

func renderPicture(layers [][]*vertex) string {
	picture := ""
	// traverse the layers starting with source repos
	for i := len(layers) - 1; i >= 0; i-- {
		l := layers[i]
		written := 0
		row, border := "", ""
		renderEdges := make([]renderEdge, 0)
		// print the row of boxed vertices
		for j := 0; j < len(l); j++ {
			v := l[j]
			sprintFunc := color.New(color.FgGreen).SprintFunc()
			if v.red {
				sprintFunc = color.New(color.FgRed).SprintFunc()
			}
			spacing := v.rowOffset - (boxWidth+2)/2 - written
			if spacing < 1 {
				spacing = 0
			}

			boxPadLeft := strings.Repeat(" ", (boxWidth-len(v.label))/2)
			boxPadRight := strings.Repeat(" ", boxWidth-len(v.label)-len(boxPadLeft))

			if v.label == "*" {
				hiddenRow := fmt.Sprintf("%s %s%s%s ", strings.Repeat(" ", spacing), boxPadLeft, "|", boxPadRight)
				border += hiddenRow
				row += hiddenRow
			} else {
				border += sprintFunc(fmt.Sprintf("%s+%s+", strings.Repeat(" ", spacing), strings.Repeat("-", boxWidth)))
				row += fmt.Sprintf("%s|%s%s%s|", strings.Repeat(" ", spacing), boxPadLeft, v, boxPadRight)
			}

			written += len(row)

			for _, u := range v.edges {
				renderEdges = append(renderEdges, renderEdge{src: v.rowOffset, dest: u.rowOffset})
			}
		}
		picture += fmt.Sprintf("%s\n%s\n%s\n", border, row, border)

		// print up to `layerVerticalSpace` rows that will contain edge drawings
		sort.Slice(renderEdges, func(i, j int) bool {
			return renderEdges[i].src < renderEdges[j].src ||
				renderEdges[i].src == renderEdges[j].src && renderEdges[i].dest < renderEdges[j].dest
		})

		for j := 0; j < layerVerticalSpace; j++ {
			row := strings.Repeat(" ", rowWidth(layers)) // TODO: calling rowWidth is expensive
			for _, re := range renderEdges {
				row = re.render(row, j, layerVerticalSpace)
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

func (re renderEdge) render(row string, vertIdx, vertDist int) string {
	setStrIdx := func(s string, i int, r rune) string {
		return s[:i] + string(r) + s[i+1:]
	}
	c := '+' // set the coordinate to "+" if there's an edge crossing
	if re.src == re.dest {
		if row[re.src] == ' ' {
			c = '|'
		}
		return setStrIdx(row, re.src, c)
	}
	const srcEdgeCenterOffset = 1                                                         // start drawing a diagonal edge one space away from the center of a node
	if vertDist > abs(re.src-re.dest) && (vertIdx > vertDist/2 || vertIdx < vertDist/2) { // vertical line
		return setStrIdx(row, (re.src+re.dest)/2, '|')
	} else if vertDist < abs(re.src-re.dest) && vertIdx == vertDist/2 { // horizontal line
		start, end := func(a, b int) (int, int) {
			if a < b {
				return a, b
			}
			return b + 1, a + 1 // weird line
		}(re.src, re.dest)
		diagCoverage := ceilDiv(vertDist, 2)
		start, end = start+diagCoverage, end-diagCoverage
		for i := start; i < end; i++ {
			row = setStrIdx(row, i, '-')
		}
		return row
	} else { // diagonal
		offset := vertIdx + srcEdgeCenterOffset
		if vertIdx > vertDist/2 {
			offset = offset + abs(re.src-re.dest) - vertDist - 1
		}
		if re.src > re.dest {
			i := re.src - offset
			if row[i] == ' ' || row[i] == '/' {
				c = '/'
			}
			return setStrIdx(row, i, c)
		} else {
			i := re.src + offset
			if row[i] == ' ' || row[i] == '\\' {
				c = '\\'
			}
			return setStrIdx(row, i, c)
		}
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

func rowWidth(layers [][]*vertex) int {
	mlw := maxLayerWidth(layers)
	return mlw*boxWidth + (mlw+1)*(padding)
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

func ceilDiv(x, y int) int {
	return int(math.Ceil(float64(x) / float64(y)))
}
