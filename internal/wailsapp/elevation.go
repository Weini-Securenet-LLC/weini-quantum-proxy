package wailsapp

type ElevationController interface {
	IsAdmin() (bool, error)
	RelaunchAsAdministrator(exe string, args []string) error
}

func EnsureElevated(exe string, args []string, controller ElevationController) (bool, error) {
	if controller == nil {
		return false, nil
	}
	isAdmin, err := controller.IsAdmin()
	if err != nil {
		return false, err
	}
	if isAdmin {
		return false, nil
	}
	if err := controller.RelaunchAsAdministrator(exe, args); err != nil {
		return false, err
	}
	return true, nil
}
