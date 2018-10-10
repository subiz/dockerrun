package main

func compareArr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareStep(a, b Step) bool {
	return a.Image == b.Image &&
		a.Dir == b.Dir &&
		a.Command == b.Command &&
		a.Shell == b.Shell &&
		compareArr(a.Volumes, b.Volumes)
}

func compareSteps(a, b []Step) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !compareStep(a[i], b[i]) {
			return false
		}
	}
	return true
}

func compareError(a, b error) bool {
	if a == nil {
		return b == nil
	}

	if b == nil {
		return false
	}

	return a.Error() == b.Error()
}
