package vcls

import "github.com/turtacn/hci-vcls/pkg/fdm"

func DefaultCapabilityMatrix() CapabilityMatrix {
	return CapabilityMatrix{
		fdm.DegradationNone: {CapabilityHA, CapabilityDRS, CapabilityFT, CapabilityVMMigration, CapabilityStoragevMotion, CapabilitySnapshots},
		fdm.DegradationZK:   {CapabilityHA, CapabilityDRS, CapabilityVMMigration},
		fdm.DegradationCFS:  {CapabilityHA, CapabilityDRS, CapabilityVMMigration},
		fdm.DegradationMySQL: {CapabilityHA},
		fdm.DegradationAll:   {},
	}
}

func ValidateCapabilityMatrix(matrix CapabilityMatrix) error {
	if _, ok := matrix[fdm.DegradationNone]; !ok {
		return ErrInvalidCapabilityMatrix
	}
	return nil
}

//Personal.AI order the ending