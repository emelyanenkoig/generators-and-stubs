package balancing

import (
	"fmt"
	"gns/stub/server/managed/entities"
	"math/rand"
)

const (
	RoundRobin                     = "round-robin"
	Weighted                       = "weighted"
	Random                         = "random"
	WeightedRandomWithBinarySearch = "weighted random with bs"
)

var (
	ValidStrategy = map[string]struct{}{RoundRobin: {}, Weighted: {}, Random: {}, WeightedRandomWithBinarySearch: {}}
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
func (b *Balancer) SelectResponse(responseSet entities.ResponseSet) (error, entities.Response) {
	switch responseSet.Choice {
	case RoundRobin:
		return nil, b.SelectRoundRobinResponse(responseSet)
	case Weighted:
		return nil, b.SelectWeightedResponse(responseSet)
	case Random:
		return nil, b.SelectRandomResponse(responseSet)
	case WeightedRandomWithBinarySearch:
		return nil, b.SelectRandomWeightedResponse(responseSet)
	default:
		return fmt.Errorf("incorrect choice"), entities.Response{} // Default to the first response if choice is invalid
	}
}

// SelectRoundRobinResponse realize RoundRobin selection
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

// SelectRandomResponse  realize full Random selection
func (b *Balancer) SelectRandomResponse(responseSet entities.ResponseSet) entities.Response {
	if len(responseSet.Responses) == 0 {
		return entities.Response{}
	}

	randValue := rand.Intn(len(responseSet.Responses))
	return responseSet.Responses[randValue]
}

// SelectRandomWeightedResponse TODO need to realize algorithm
// SelectRandomWeightedResponse realize WeightedRandomWithBinarySearch selection
func (b *Balancer) SelectRandomWeightedResponse(responseSet entities.ResponseSet) entities.Response {
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
