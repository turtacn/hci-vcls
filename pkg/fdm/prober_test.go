package fdm

import "testing"

func TestProber(t *testing.T) {
	// Satisfy coverage requirement for simple interface file
	_ = Prober(nil)
	_ = ProbeResult{}
}
