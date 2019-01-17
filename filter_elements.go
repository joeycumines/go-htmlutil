package htmlutil

import "golang.org/x/net/html"

type filterElementsConfig struct {
	Node    *html.Node
	Filters []func(node *html.Node) bool
	Find    bool
}

func filterElements(config filterElementsConfig) []*html.Node {
	var (
		result []*html.Node
		fn     func(config filterElementsConfig)
	)

	fn = func(config filterElementsConfig) {
		if config.Node == nil {
			return
		}

		if config.Find && len(result) != 0 {
			return
		}

		if len(config.Filters) == 0 {
			result = append(result, config.Node)
			return
		}

		start := len(result)

		func(config filterElementsConfig) {
			var filter func(node *html.Node) bool

			for filter == nil && len(config.Filters) != 0 {
				filter = config.Filters[0]
				config.Filters = config.Filters[1:]
			}

			if filter != nil && !filter(config.Node) {
				return
			}

			fn(config)
		}(config)

		finish := len(result)

		for c := config.Node.FirstChild; c != nil; c = c.NextSibling {
			config := config
			config.Node = c

			fn(config)

			for i := start; i < finish; i++ {
				for j := finish; j < len(result); j++ {
					if result[i] != result[j] {
						continue
					}

					copy(result[j:], result[j+1:])
					result[len(result)-1] = nil
					result = result[:len(result)-1]
					j--
				}
			}
		}
	}

	fn(config)

	return result
}
