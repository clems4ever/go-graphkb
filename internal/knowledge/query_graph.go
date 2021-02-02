package knowledge

import (
	"errors"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/clems4ever/go-graphkb/internal/utils"
)

// ErrVariableNotFound error thrown when a variable does not exist
var ErrVariableNotFound = errors.New("Unable to find variable")

// RelationDirection the direction of a relation
type RelationDirection int

const (
	// Left relation
	Left RelationDirection = iota
	// Right relation
	Right RelationDirection = iota
	// Either there is a relation but we don't know in which direction
	Either RelationDirection = iota
	// Both there is a relation in both directions
	Both RelationDirection = iota
)

// QueryNode represent a node and its constraints
type QueryNode struct {
	Labels []string
	// Constraint expressions
	Constraints AndOrExpression
}

// QueryRelation represent a relation and its constraints
type QueryRelation struct {
	Labels []string
	// Constraint expressions
	Constraints AndOrExpression

	LeftIdx   int
	RightIdx  int
	Direction RelationDirection
}

// VariableType represent the type of a variable in the cypher query.
type VariableType int

const (
	// NodeType variable of type node
	NodeType VariableType = iota
	// RelationType variable of type relation
	RelationType VariableType = iota
)

// TypeAndIndex type and index of a variable from the cypher query
type TypeAndIndex struct {
	Type  VariableType
	Index int
}

// QueryGraph the representation of a query graph. This structure helps create the relations between nodes to facilitate SQL translation and projections
type QueryGraph struct {
	Nodes     []QueryNode
	Relations []QueryRelation

	VariablesIndex map[string]TypeAndIndex
}

// NewQueryGraph create an instance of a query graph
func NewQueryGraph() QueryGraph {
	return QueryGraph{
		Nodes:          []QueryNode{},
		Relations:      []QueryRelation{},
		VariablesIndex: make(map[string]TypeAndIndex),
	}
}

// PushNode push a node into the registry
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

// PushRelation push a relation into the registry
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

// FindVariable find a variable by its name
func (qg *QueryGraph) FindVariable(name string) (TypeAndIndex, error) {
	v, ok := qg.VariablesIndex[name]
	if !ok {
		return TypeAndIndex{}, ErrVariableNotFound
	}
	return v, nil
}

// GetNodeByID get a node by its id
func (qg *QueryGraph) GetNodeByID(idx int) (*QueryNode, error) {
	if idx >= len(qg.Nodes) {
		return nil, fmt.Errorf("Index provided to find node is invalid")
	}
	return &qg.Nodes[idx], nil
}
