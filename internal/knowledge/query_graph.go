package knowledge

import (
	"errors"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/clems4ever/go-graphkb/internal/utils"
)

var ErrVariableNotFound = errors.New("Unable to find variable")

// QueryNode represent a node and its constraints
type QueryNode struct {
	Labels []string
	// Constraint expressions
	Constraints AndOrExpression
}

type RelationDirection int

const (
	// Left relation
	Left RelationDirection = iota
	// Right relation
	Right RelationDirection = iota
	// There is a relation but we don't know in which direction
	Either RelationDirection = iota
	// There is a relation in both directions
	Both RelationDirection = iota
)

// QueryRelation represent a relation and its constraints
type QueryRelation struct {
	Labels []string
	// Constraint expressions
	Constraints AndOrExpression

	LeftIdx   int
	RightIdx  int
	Direction RelationDirection
}

type VariableType int

const (
	NodeType     VariableType = iota
	RelationType VariableType = iota
)

type TypeAndIndex struct {
	Type  VariableType
	Index int
}

type QueryGraph struct {
	Nodes     []QueryNode
	Relations []QueryRelation

	VariablesIndex map[string]TypeAndIndex
}

func NewQueryGraph() QueryGraph {
	return QueryGraph{
		Nodes:          []QueryNode{},
		Relations:      []QueryRelation{},
		VariablesIndex: make(map[string]TypeAndIndex),
	}
}

func (qg *QueryGraph) PushNode(q query.QueryNodePattern) (*QueryNode, int, error) {
	// If pattern comes with a variable name, search in the index if it does not already exist
	if q.Variable != "" {
		typeAndIndex, ok := qg.VariablesIndex[q.Variable]

		// If found, returns the node
		if ok {
			if typeAndIndex.Type != NodeType {
				return nil, -1, fmt.Errorf("Variable '%s' is assigned to a different type", q.Variable)
			}

			n := qg.Nodes[typeAndIndex.Index]
			if !utils.AreStringSliceElementsEqual(n.Labels, q.Labels) && q.Labels != nil {
				return nil, -1, fmt.Errorf("Variable '%s' already defined with a different type", q.Variable)
			}
			return &n, typeAndIndex.Index, nil
		}
	}

	qn := QueryNode{Labels: q.Labels}
	newIdx := len(qg.Nodes)

	qg.Nodes = append(qg.Nodes, qn)
	if q.Variable != "" {
		qg.VariablesIndex[q.Variable] = TypeAndIndex{
			Type:  NodeType,
			Index: newIdx,
		}
	}

	return &qn, newIdx, nil
}

func (qg *QueryGraph) PushRelation(q query.QueryRelationshipPattern, leftIdx, rightIdx int) (*QueryRelation, int, error) {
	var varName string
	var labels []string

	if q.RelationshipDetail != nil {
		varName = q.RelationshipDetail.Variable
		labels = q.RelationshipDetail.Labels
	}

	// If pattern comes with a variable name, search in the index if it does not already exist
	if varName != "" {
		typeAndIndex, ok := qg.VariablesIndex[varName]
		// If found, returns the node
		if ok {
			if typeAndIndex.Type != RelationType {
				return nil, -1, fmt.Errorf("Variable '%s' is assigned to a different type", varName)
			}
			r := qg.Relations[typeAndIndex.Index]
			if !utils.AreStringSliceElementsEqual(r.Labels, labels) {
				return nil, -1, fmt.Errorf("Variable '%s' already defined with a different type", varName)
			}
			return &r, typeAndIndex.Index, nil
		}
	}

	if leftIdx >= len(qg.Nodes) {
		return nil, -1, fmt.Errorf("Cannot push relation bound to an unexisting node")
	}

	if rightIdx >= len(qg.Nodes) {
		return nil, -1, fmt.Errorf("Cannot push relation bound to an unexisting node")
	}

	var direction RelationDirection
	if !q.LeftArrow && !q.RightArrow {
		direction = Either
	} else if q.LeftArrow && q.RightArrow {
		direction = Both
	} else if q.LeftArrow {
		direction = Left
	} else if q.RightArrow {
		direction = Right
	} else {
		return nil, -1, fmt.Errorf("Unable to detection the direction of the relation")
	}

	qr := QueryRelation{
		Labels:    labels,
		LeftIdx:   leftIdx,
		RightIdx:  rightIdx,
		Direction: direction,
	}
	newIdx := len(qg.Relations)

	qg.Relations = append(qg.Relations, qr)
	if varName != "" {
		qg.VariablesIndex[varName] = TypeAndIndex{
			Type:  RelationType,
			Index: newIdx,
		}
	}

	return &qr, newIdx, nil
}

func (qg *QueryGraph) FindVariable(name string) (TypeAndIndex, error) {
	v, ok := qg.VariablesIndex[name]
	if !ok {
		return TypeAndIndex{}, ErrVariableNotFound
	}
	return v, nil
}

func (gq *QueryGraph) FindNode(idx int) (*QueryNode, error) {
	if idx >= len(gq.Nodes) {
		return nil, fmt.Errorf("Index provided to find node is invalid")
	}
	return &gq.Nodes[idx], nil
}
