package main

import "fmt"

// Pitfall: Modifications not visible to caller
func addElementWrong(s []int) {
	s = append(s, 99)
	fmt.Printf("  Inside function: %v (len=%d, cap=%d)\n", s, len(s), cap(s))
}

// Solution 1: Return the modified slice (idiomatic Go)
func addElementReturn(s []int) []int {
	s = append(s, 99)
	return s
}

// Solution 2: Pass pointer to slice (like Rust borrow)
func addElementPointer(s *[]int) {
	*s = append(*s, 99)
}

// Demonstration: append is NOT immutable
func demonstrateNotImmutable() {
	fmt.Println("\n=== Append is NOT Immutable ===")

	// Create slice with extra capacity
	s1 := make([]int, 2, 5)
	s1[0], s1[1] = 10, 20

	// s2 shares the same underlying array
	s2 := s1[:2]

	fmt.Printf("Before append:\n")
	fmt.Printf("  s1: %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))
	fmt.Printf("  s2: %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))

	// Append within capacity - modifies underlying array in-place
	s1 = append(s1, 30)

	fmt.Printf("After s1 = append(s1, 30):\n")
	fmt.Printf("  s1: %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))
	fmt.Printf("  s2: %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))
	fmt.Printf("  s2 with extended view: %v\n", s2[:cap(s2)])
	fmt.Println("  ⚠️  s2 can see the mutation! Append modified the shared array.")
}

// Demonstration: When append reallocates
func demonstrateReallocation() {
	fmt.Println("\n=== When Append Reallocates ===")

	s1 := []int{1, 2, 3} // Small capacity
	s2 := s1

	fmt.Printf("Before append:\n")
	fmt.Printf("  s1: %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))
	fmt.Printf("  s2: %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))

	// Append beyond capacity - triggers reallocation
	s1 = append(s1, 4, 5, 6, 7, 8)

	fmt.Printf("After s1 = append(s1, 4, 5, 6, 7, 8):\n")
	fmt.Printf("  s1: %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))
	fmt.Printf("  s2: %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))
	fmt.Println("  ℹ️  s1 now points to a new array; s2 still points to the old one.")
}

func main() {
	fmt.Println("=== Slice Modification Pitfalls ===")

	// Pitfall: Changes not visible
	fmt.Println("\n1. PITFALL - Passing slice by value:")
	slice1 := []int{1, 2, 3}
	fmt.Printf("Before: %v (len=%d, cap=%d)\n", slice1, len(slice1), cap(slice1))

	// passing the metadata (pointer, len, cap) by value
	// so any reallocation inside the function won't affect the caller's slice
	addElementWrong(slice1)

	fmt.Printf("After:  %v (len=%d, cap=%d)\n", slice1, len(slice1), cap(slice1))
	fmt.Println("  ❌ Change NOT visible - append reallocated inside function")

	// Solution 1: Return the slice
	fmt.Println("\n2. SOLUTION 1 - Return the modified slice (idiomatic):")
	slice2 := []int{1, 2, 3}
	fmt.Printf("Before: %v\n", slice2)
	slice2 = addElementReturn(slice2)
	fmt.Printf("After:  %v\n", slice2)
	fmt.Println("  ✅ Change visible")

	// Solution 2: Pass pointer to slice
	fmt.Println("\n3. SOLUTION 2 - Pass pointer to slice (like Rust &mut):")
	slice3 := []int{1, 2, 3}
	fmt.Printf("Before: %v\n", slice3)
	addElementPointer(&slice3) // Note: &slice3, not slice3
	fmt.Printf("After:  %v\n", slice3)
	fmt.Println("  ✅ Change visible")

	// Demonstrate append is not immutable
	demonstrateNotImmutable()

	// Demonstrate reallocation
	demonstrateReallocation()

	fmt.Println("\n=== Key Takeaways ===")
	fmt.Println("• Slices are passed by value (copy of slice header)")
	fmt.Println("• Append may or may not reallocate depending on capacity")
	fmt.Println("• Append is NOT immutable - it mutates when there's capacity")
	fmt.Println("• Use 's = append(s, x)' pattern or pass *[]T for modifications")
}
