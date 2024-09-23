package control

import (
	"math/rand"
)

// Select a response based on the strategy ("round-robin" or "weight")
func (cs *ControlServer) SelectResponse(responseSet ResponseSet) Response {
	switch responseSet.Choice {
	case "round-robin":
		return cs.SelectRoundRobinResponse(responseSet)
	case "weight":
		return cs.SelectWeightedResponse(responseSet)
	default:
		return responseSet.Responses[0] // Default to the first response if choice is invalid
	}
}

// Round-robin selection
func (cs *ControlServer) SelectRoundRobinResponse(responseSet ResponseSet) Response {
	if len(responseSet.Responses) == 0 {
		return Response{}
	}

	// Инициализация индекса, если он еще не существует для данного выбора
	if _, exists := cs.RRobinIndex[responseSet.Choice]; !exists {
		cs.RRobinIndex[responseSet.Choice] = 0
	}

	// Получаем текущий индекс
	index := cs.RRobinIndex[responseSet.Choice]

	// Обновляем индекс для следующего вызова
	cs.RRobinIndex[responseSet.Choice] = (index + 1) % len(responseSet.Responses)

	// Возвращаем выбранный ответ
	return responseSet.Responses[index]
}

// Weighted random selection
func (cs *ControlServer) SelectWeightedResponse(responseSet ResponseSet) Response {
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
