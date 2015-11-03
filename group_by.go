package kapacitor

import (
	"sort"

	"github.com/influxdb/kapacitor/models"
	"github.com/influxdb/kapacitor/pipeline"
	"github.com/influxdb/kapacitor/tick"
)

type GroupByNode struct {
	node
	g             *pipeline.GroupByNode
	dimensions    []string
	allDimensions bool
}

// Create a new GroupByNode which splits the stream dynamically based on the specified dimensions.
func newGroupByNode(et *ExecutingTask, n *pipeline.GroupByNode) (*GroupByNode, error) {
	gn := &GroupByNode{
		node: node{Node: n, et: et},
		g:    n,
	}
	gn.node.runF = gn.runGroupBy

DIMS:
	for _, dim := range n.Dimensions {
		switch d := dim.(type) {
		case string:
			gn.dimensions = append(gn.dimensions, d)
		case *tick.StarNode:
			gn.allDimensions = true
			break DIMS
		}
	}
	sort.Strings(gn.dimensions)
	return gn, nil
}

func (g *GroupByNode) runGroupBy() error {
	switch g.Wants() {
	case pipeline.StreamEdge:
		for pt, ok := g.ins[0].NextPoint(); ok; pt, ok = g.ins[0].NextPoint() {
			var tags map[string]string
			if g.allDimensions {
				tags = pt.Tags
			} else {
				tags = make(map[string]string, len(g.dimensions))
				for _, dim := range g.dimensions {
					tags[dim] = pt.Tags[dim]
				}
			}
			pt.Tags = tags
			pt.Group = models.TagsToGroupID(g.dimensions, pt.Tags)
			for _, child := range g.outs {
				err := child.CollectPoint(pt)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}