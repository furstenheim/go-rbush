package go_rbush

import (
	"sort"
	"math"
)
// sort a slice so that items come in groups of unsorted arrays of length n with groups sorted between each other
// selection algorithm + binary divide and conquer
func FloydRivestBuckets (array sort.Interface, n, left, right int) {
	// It would be nice to use multiple goroutines. Main problem is swap(left) which is right from another slice
	s := stack([]int{left, right})
	var mid int
	for len(s) > 0 {
		s, right = s.pop()
		s, left = s.pop()
		if (right - left <= n) {
			continue
		}
		// max is to avoid the case where left - right < 2n
		mid = left + max((right - left) / n / 2 , 1) * n
		FloydRivestSelect(array, mid, left, right)
		s = s.push(left)
		s = s.push(mid)
		s = s.push(mid)
		s = s.push(right)

	}
}
// left is the left index for the interval
// right is the right index for the interval
// k is the desired index value, where array[k] is the k+1 smallest element
// when left = 0
func FloydRivestSelect (array sort.Interface, k, left, right int) {
	for (right > left) {
		if (right - left > 600) {
			var n = float64(right - left + 1)
			var kf = float64(k)
			var m = float64(k - left + 1)
			var z = math.Log(n)
			var s = 0.5 * math.Exp(2 * z / 3)
			sign := float64(1)
			if  m - n / 2 < 0 {
				sign = -1
			}
			var sd = 0.5 * math.Sqrt(z * s * (n - s) / n) * sign
			var newLeft = max(left, int(math.Floor(kf - m * s / n + sd)))
			var newRight = min(right, int(math.Floor(kf + (n - m) * s / n + sd)))
			FloydRivestSelect(array, k, newLeft, newRight)
		}

		var i = left
		var j = right
		array.Swap(left, k)
		// in the original algorithm array[k] is stored to a value. To use golangs sort interface we need to keep track of the changes for the index
		// we define it as right because in the first iteration of for i<j it will be changed
		pointIndex := right
		if array.Less(k, right) {
			array.Swap(left, right)
			pointIndex = left
		}

		for i < j {
			// pointIndex is swapped only once in the first iteration. Later it will either be bigger (if left) or smaller (if right)
			array.Swap(i, j)
			i++
			j--
			for i < array.Len() && array.Less(i, pointIndex) {
				i++
			}
			for j >= 0 && array.Less(pointIndex, j) {
				j--
			}
		}
		if !array.Less(left, pointIndex) && !array.Less(pointIndex, left) {
			array.Swap(left, j)
		} else {
			j++
			array.Swap(j, right)
		}
		if (j <= k) {
			left = j + 1
		}
		if (k <= j) {
			right = j - 1
		}
	}
}

func min (a, b int) int{
	if a < b {
		return a
	}
	return b
}


func max (a, b int) int {
	if a > b {
		return a
	}
	return b
}

type stack []int

func (s stack) push (v int) stack {
	return append(s, v)
}
func (s stack) pop () (stack, int) {
	l := len(s)
	return s[:l-1], s[l-1]
}