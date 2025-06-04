package llmangoagents

import "fmt"

// DEPENDENCY GRAPH SYSTEM FOR COMPILATION
// =====================================

type DependencyNodeType string

const (
	NodeTypeAgent    DependencyNodeType = "agent"
	NodeTypeWorkflow DependencyNodeType = "workflow"
	NodeTypeTool     DependencyNodeType = "tool"
)

type DependencyNode struct {
	ID           string // unique identifier (agent/workflow/tool UID)
	Type         DependencyNodeType
	Dependencies []string    // list of IDs this node depends on
	Dependents   []string    // list of IDs that depend on this node
	InDegree     int         // number of incoming edges (dependencies)
	Compiled     bool        // whether this node has been compiled
	CompileData  interface{} // stores the compiled result
}

type DependencyGraph struct {
	Nodes            map[string]*DependencyNode
	CompilationOrder []string   // result of topological sort
	CircularDeps     [][]string // detected circular dependencies
}

// BuildDependencyGraph creates a dependency graph from system inputs
func BuildDependencyGraph(inputs SystemInputList) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// PHASE 1: Create tool registry (tools are external dependencies, not part of dependency graph)
	toolRegistry := make(map[string]bool)
	for _, tool := range inputs.Tools {
		toolRegistry[tool.UID] = true
	}

	// PHASE 2: Create agent and workflow nodes with their internal system dependencies
	for _, agent := range inputs.Agents {
		deps := extractAgentDependencies(agent, toolRegistry)
		graph.addNode(agent.UID, NodeTypeAgent, deps)
	}

	for _, workflow := range inputs.Workflows {
		deps := extractWorkflowDependencies(workflow, toolRegistry)
		graph.addNode(workflow.UID, NodeTypeWorkflow, deps)
	}

	// PHASE 3: Build bidirectional edges and calculate in-degrees
	if err := graph.buildEdges(); err != nil {
		return nil, err
	}

	// PHASE 4: Detect circular dependencies
	if cycles := graph.detectCircularDependencies(); len(cycles) > 0 {
		graph.CircularDeps = cycles
		return graph, fmt.Errorf("circular dependencies detected: %v", cycles)
	}

	// PHASE 5: Generate compilation order using Khan's algorithm
	order, err := graph.topologicalSort()
	if err != nil {
		return nil, err
	}
	graph.CompilationOrder = order

	return graph, nil
}

func (dg *DependencyGraph) addNode(id string, nodeType DependencyNodeType, deps []string) {
	dg.Nodes[id] = &DependencyNode{
		ID:           id,
		Type:         nodeType,
		Dependencies: deps,
		Dependents:   []string{},
		InDegree:     len(deps),
		Compiled:     false,
	}
}

func (dg *DependencyGraph) buildEdges() error {
	for nodeID, node := range dg.Nodes {
		for _, depID := range node.Dependencies {
			depNode, exists := dg.Nodes[depID]
			if !exists {
				return fmt.Errorf("dependency '%s' not found for node '%s'", depID, nodeID)
			}

			// Add bidirectional reference
			depNode.Dependents = append(depNode.Dependents, nodeID)
		}
	}
	return nil
}

// KHAN'S ALGORITHM IMPLEMENTATION
func (dg *DependencyGraph) topologicalSort() ([]string, error) {
	// Create a copy of in-degrees for manipulation
	inDegrees := make(map[string]int)
	for id, node := range dg.Nodes {
		inDegrees[id] = node.InDegree
	}

	// Initialize queue with nodes having no dependencies
	queue := []string{}
	for id, degree := range inDegrees {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	compilationOrder := []string{}

	// Khan's algorithm main loop
	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]
		compilationOrder = append(compilationOrder, current)

		// Reduce in-degree of all dependent nodes
		currentNode := dg.Nodes[current]
		for _, dependentID := range currentNode.Dependents {
			inDegrees[dependentID]--

			// If dependent now has no incoming edges, add to queue
			if inDegrees[dependentID] == 0 {
				queue = append(queue, dependentID)
			}
		}
	}

	// Check if all nodes were processed (no cycles)
	if len(compilationOrder) != len(dg.Nodes) {
		return nil, fmt.Errorf("circular dependency detected - could not compile all nodes")
	}

	return compilationOrder, nil
}

// DEPENDENCY EXTRACTION LOGIC
func extractAgentDependencies(agent Agent, toolRegistry map[string]bool) []string {
	deps := []string{}

	// Agent-level internal dependencies (all string IDs before compilation)
	// SubAgents that become internal agent tools
	deps = append(deps, agent.SubAgents...)

	// SubWorkflows that become internal agent tools
	deps = append(deps, agent.SubWorkflows...)

	// agent.Tools contains ONLY external tools (not dependencies)

	return deps
}

func extractWorkflowDependencies(workflow Workflow, toolRegistry map[string]bool) []string {
	deps := []string{}

	for _, step := range workflow.Steps {
		// Lead agent dependency (string ID)
		deps = append(deps, step.Agent)

		// Step-level SubAgent dependencies ([]string)
		deps = append(deps, step.SubAgents...)
	}

	return removeDuplicates(deps)
}

// CIRCULAR DEPENDENCY DETECTION
func (dg *DependencyGraph) detectCircularDependencies() [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range dg.Nodes {
		if !visited[nodeID] {
			if cyclePath := dfsDetectCycle(dg, nodeID, visited, recStack, []string{}); len(cyclePath) > 0 {
				cycles = append(cycles, cyclePath)
			}
		}
	}

	return cycles
}

func dfsDetectCycle(dg *DependencyGraph, nodeID string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	node := dg.Nodes[nodeID]
	for _, depID := range node.Dependencies {
		if !visited[depID] {
			if cycle := dfsDetectCycle(dg, depID, visited, recStack, path); len(cycle) > 0 {
				return cycle
			}
		} else if recStack[depID] {
			// Found cycle - return the cycle path
			cycleStart := -1
			for i, id := range path {
				if id == depID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], depID)
			}
		}
	}

	recStack[nodeID] = false
	return []string{}
}

// UTILITY FUNCTIONS
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
