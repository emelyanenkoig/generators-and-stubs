package balancing

import (
	"gns/stub/server/managed/entities"
	"math/rand"
)

type Balancer struct {
	RRobinIndex int
}

func InitBalancer() *Balancer {
	return &Balancer{
		RRobinIndex: 0,
	}

}

// SelectResponse based on the strategy ("round-robin", "weight", "random")
func (b *Balancer) SelectResponse(responseSet entities.ResponseSet) entities.Response {
	switch responseSet.Choice {
	case "round-robin":
		return b.SelectRoundRobinResponse(responseSet)
	case "weight":
		return b.SelectWeightedResponse(responseSet)
	case "random":
		return b.SelectRandomResponse(responseSet)
	default:
		return responseSet.Responses[0] // Default to the first response if choice is invalid
	}
}

// SelectRoundRobinResponse realize Round-robin selection
func (b *Balancer) SelectRoundRobinResponse(responseSet entities.ResponseSet) entities.Response {
	if len(responseSet.Responses) == 0 {
		return entities.Response{}
	}

	curRrIndex := b.RRobinIndex
	b.RRobinIndex = (curRrIndex + 1) % len(responseSet.Responses)

	return responseSet.Responses[curRrIndex]
}

// SelectWeightedResponse realize Weighted random selection
func (b *Balancer) SelectWeightedResponse(responseSet entities.ResponseSet) entities.Response {
	if len(responseSet.Responses) == 0 {
		return entities.Response{}
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

// SelectRandomResponse  realize full random selection
func (b *Balancer) SelectRandomResponse(responseSet entities.ResponseSet) entities.Response {
	if len(responseSet.Responses) == 0 {
		return entities.Response{}
	}

	randValue := rand.Intn(len(responseSet.Responses))
	return responseSet.Responses[randValue]
}
