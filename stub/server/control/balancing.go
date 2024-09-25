package control

import (
	"math/rand"
)

type Balancer struct {
	RRobinIndex map[string]int
}

func (b *Balancer) InitBalancer() {
	b.RRobinIndex = make(map[string]int)
}

// Select a response based on the strategy ("round-robin" or "weight")
func (b *Balancer) SelectResponse(responseSet ResponseSet) Response {
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

// Round-robin selection
func (b *Balancer) SelectRoundRobinResponse(responseSet ResponseSet) Response {
	if len(responseSet.Responses) == 0 {
		return Response{}
	}

	// Инициализация индекса, если он еще не существует для данного выбора
	if _, exists := b.RRobinIndex[responseSet.Choice]; !exists {
		b.RRobinIndex[responseSet.Choice] = 0
	}
	index := b.RRobinIndex[responseSet.Choice]

	// Обновляем индекс для следующего вызова
	b.RRobinIndex[responseSet.Choice] = (index + 1) % len(responseSet.Responses)

	return responseSet.Responses[index]
}

// Weighted random selection
func (b *Balancer) SelectWeightedResponse(responseSet ResponseSet) Response {
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

func (b *Balancer) SelectRandomResponse(responseSet ResponseSet) Response {
	if len(responseSet.Responses) == 0 {
		return Response{}
	}

	randValue := rand.Intn(len(responseSet.Responses))
	return responseSet.Responses[randValue]
}
