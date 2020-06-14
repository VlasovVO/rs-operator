package controller

import (
	"myrs-operator/pkg/controller/myrs"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, myrs.Add)
}
