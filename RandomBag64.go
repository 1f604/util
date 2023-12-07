package util

// custom error types

type RandomBagEmptyError struct{}

func (e RandomBagEmptyError) Error() string {
	return "RandomBag Error: Cannot pop because there are no elements in the bag."
}

type RandomBag64 struct {
	arr []uint64
}

func (rb *RandomBag64) Size() int {
	return len(rb.arr)
}

// Removes from array and swaps last element into it
func (rb *RandomBag64) PopRandom() (uint64, error) {
	// fmt.Println("Bag initial:", rb.arr)
	// check if something can be popped
	if len(rb.arr) == 0 {
		return 0, RandomBagEmptyError{}
	}

	index, err := Crypto_Randint(len(rb.arr))
	if err != nil {
		return 0, err
	}

	elem := rb.arr[index]
	n := len(rb.arr)
	// swap it with the last element
	rb.arr[index] = rb.arr[n-1]
	// resize the array
	rb.arr = rb.arr[:n-1]
	// fmt.Println("Returned element:", elem)
	// fmt.Println("Bag final:", rb.arr)
	return elem, nil
}

// Push should always succeed
func (rb *RandomBag64) Push(item uint64) {
	rb.arr = append(rb.arr, item)
}

// The RandomBag steals the slice that you pass to it. You should not use the slice anywhere afterwards.
func CreateRandomBagFromSlice(items []uint64) *RandomBag64 {
	return &RandomBag64{
		arr: items,
	}
}
