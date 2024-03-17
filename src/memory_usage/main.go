package main

// DynamicArray represents a resizable array.
type DynamicArray struct {
	array    *[]int // Pointer to an integer slice
	size     int    // Number of elements in the array
	capacity int    // Maximum number of elements the array can hold
}

// NewDynamicArray initializes a new dynamic array.
func NewDynamicArray(size int, capacity int) *DynamicArray {
	arr := make([]int, size, capacity)
	return &DynamicArray{array: &arr, size: 0, capacity: capacity}
}

// AddElement adds an element to the end of the array.
func (da *DynamicArray) AddElement(element int) {
	// Check if the array needs resizing
	if da.size == da.capacity {
		da.Resize(2 * da.capacity)
	}

	// Add the element
	(*da.array)[da.size] = element
	da.size++
}

func main() {
	arr := NewDynamicArray(3, 10)
	println(arr.capacity)
	// length of arr.array
	println(len(*arr.array))
}
