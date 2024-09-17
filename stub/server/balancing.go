package server

import "math/rand"

// Select a response based on the strategy ("round-robin" or "weight")
func (c *ControlServer) SelectResponse(responseSet ResponseSet) Response {
	switch responseSet.Choice {
	case "round-robin":
		return c.SelectRoundRobinResponse(responseSet)
	case "weight":
		return c.SelectWeightedResponse(responseSet)
	default:
		return responseSet.Responses[0] // Default to the first response if choice is invalid
	}
}

// Round-robin selection
func (c *ControlServer) SelectRoundRobinResponse(responseSet ResponseSet) Response {
	if len(responseSet.Responses) == 0 {
		return Response{}
	}

	// Lock only around the critical section that modifies the shared RRobinIndex
	index := c.RRobinIndex[responseSet.Choice]                                   // Read current index
	c.RRobinIndex[responseSet.Choice] = (index + 1) % len(responseSet.Responses) // Update index

	return responseSet.Responses[index] // Return selected response
}

// Weighted random selection
func (c *ControlServer) SelectWeightedResponse(responseSet ResponseSet) Response {
	if len(responseSet.Responses) == 0 {
		return Response{}
	}

	totalWeight := 0
	for _, response := range responseSet.Responses {
		totalWeight += response.Weight
	}

	randValue := rand.Intn(totalWeight)
	cumulativeWeight := 0
	for _, response := range responseSet.Responses {
		cumulativeWeight += response.Weight
		if randValue < cumulativeWeight {
			return response
		}
	}

	return responseSet.Responses[0]
}
