// pipe.go implements Pipe for node IPC, it is designed to stream data between
// nodes recursively so that data is consistent between nodes
package sdfs

type Pipe struct {
	remainingNodes []string
}

// Push adds a node to Pipe
func (p *Pipe) Push(node string) {
	p.remainingNodes = append(p.remainingNodes, node)
}

// Pop gets a node from pipe, returns empty string if the Pipe is empty
func (p *Pipe) Pop() (result string) {
	if len(p.remainingNodes) > 0 {
		result = p.remainingNodes[len(p.remainingNodes)-1]
		p.remainingNodes = p.remainingNodes[:len(p.remainingNodes)-1]
		return
	} else {
		return ""
	}
}
