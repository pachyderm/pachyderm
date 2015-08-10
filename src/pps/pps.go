package pps

func GetNameToNode(pipeline *Pipeline) map[string]*Node {
	m := make(map[string]*Node)
	for name, element := range pipeline.NameToElement {
		if element.Node != nil {
			m[name] = element.Node
		}
	}
	return m
}

func GetNameToDockerService(pipeline *Pipeline) map[string]*DockerService {
	m := make(map[string]*DockerService)
	for name, element := range pipeline.NameToElement {
		if element.DockerService != nil {
			m[name] = element.DockerService
		}
	}
	return m
}
