package domain

// CountToScore converts a raw count of observed updates to a velocity score
func CountToScore(count int32) int32 {
	if count == 0 {
		return 0
	}

	if count > 0 && count < 6 {
		return 1
	}

	if count >= 6 && count < 11 {
		return 2
	}

	if count >= 11 && count < 16 {
		return 3
	}
	if count >= 16 && count < 21 {
		return 4
	}

	return 5
}
