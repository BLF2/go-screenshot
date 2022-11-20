package util

import "container/list"

func SplitMap(inputMap map[string]string, cap int) *list.List {

	if inputMap == nil || len(inputMap) == 0 {
		return list.New()
	}

	length := len(inputMap)
	maxCnt := length / cap

	if maxCnt*cap < length {
		maxCnt++
	}

	var listMap list.List

	currentCnt := 0
	subMap := make(map[string]string, cap)
	for key := range inputMap {
		subMap[key] = inputMap[key]
		currentCnt++
		if currentCnt == cap {
			listMap.PushBack(subMap)
			currentCnt = 0
			subMap = make(map[string]string, cap)
		}
	}
	//说明除不尽有剩余
	if currentCnt > 0 && currentCnt < cap {
		listMap.PushBack(subMap)
	}

	return &listMap
}
