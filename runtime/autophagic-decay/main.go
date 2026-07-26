package main

import (
	"fmt"
	"math"
)

// BoundaryObject represents a claim traversing the Cybot network
type BoundaryObject struct {
	ID                 string
	Content            string
	GenerationalDepth  int
	BaseUncertainty    float64
}

// CalculateEffectiveUncertainty applies the Epistemic Decay formula
// U_eff = 1 - (1 - U_base) * e^(-lambda * D)
func CalculateEffectiveUncertainty(obj BoundaryObject, lambda float64) float64 {
	if obj.GenerationalDepth == 0 {
		return obj.BaseUncertainty
	}
	decayFactor := math.Exp(-lambda * float64(obj.GenerationalDepth))
	return 1.0 - ((1.0 - obj.BaseUncertainty) * decayFactor)
}

func main() {
	fmt.Println("=== PROPOSED CANON: Autophagic Decay (α) Simulator ===")
	fmt.Println("Demonstrating epistemic collapse in recursive local-first RAG loops.")
	fmt.Println()

	// Decay constant (tunable parameter for half-life)
	// lambda = 0.2 means approx 3.4 generations until half the remaining certainty is lost
	lambda := 0.2 
	generations := 10

	// D=0 is a primary source.
	rootSource := BoundaryObject{
		ID:                "GEN-0",
		Content:           "Root empirical observation (D=0)",
		GenerationalDepth: 0,
		BaseUncertainty:   0.10, // 10% uncertainty at the root
	}

	fmt.Printf("Decay Constant (λ): %.2f\n", lambda)
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("%-10s | %-5s | %-12s | %-12s\n", "Generation", "Depth", "Base Uncert", "Effective Uncert")
	fmt.Println("---------------------------------------------------------")

	currentObj := rootSource

	for d := 0; d <= generations; d++ {
		currentObj.GenerationalDepth = d
		currentObj.ID = fmt.Sprintf("GEN-%d", d)
		
		effU := CalculateEffectiveUncertainty(currentObj, lambda)
		
		fmt.Printf("%-10s | %-5d | %.4f       | %.4f\n", currentObj.ID, currentObj.GenerationalDepth, currentObj.BaseUncertainty, effU)

		// Simulate the RAG feedback loop: The Cybot reads GEN-d to produce GEN-(d+1)
		// The base uncertainty remains 0.10 because the Cybot doesn't "know" it's degrading,
		// but the geometric law forces the Effective Uncertainty to inflate.
	}

	fmt.Println("---------------------------------------------------------")
	fmt.Println("\n=== Simulator Observation ===")
	fmt.Println("Notice how the Cybot's internal base uncertainty never changes (the AI remains confident),")
	fmt.Println("but the Effective Uncertainty asymptotically approaches 1.0 (Total Unknown) as D grows.")
	fmt.Println("To restore certainty, the Cybot must mathematically route back to a D=0 anchor.")
}
