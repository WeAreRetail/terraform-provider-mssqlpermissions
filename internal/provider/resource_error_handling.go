package provider

// ErrorHandlingResult represents the decision of how to handle an error in a Read method
type ErrorHandlingResult struct {
	ShouldRemoveFromState bool
	ShouldAddError        bool
	ErrorMessage          string
}

// HandleUserReadError analyzes an error from GetUser and determines the appropriate action
// This function encapsulates the logic for handling "user not found" vs other errors
func HandleUserReadError(err error) ErrorHandlingResult {
	if err == nil {
		return ErrorHandlingResult{
			ShouldRemoveFromState: false,
			ShouldAddError:        false,
		}
	}

	if err.Error() == "user not found" {
		return ErrorHandlingResult{
			ShouldRemoveFromState: true,
			ShouldAddError:        false,
		}
	}

	return ErrorHandlingResult{
		ShouldRemoveFromState: false,
		ShouldAddError:        true,
		ErrorMessage:          "Error getting user",
	}
}

// HandleDatabaseRoleReadError analyzes an error from GetDatabaseRole and determines the appropriate action
func HandleDatabaseRoleReadError(err error) ErrorHandlingResult {
	if err == nil {
		return ErrorHandlingResult{
			ShouldRemoveFromState: false,
			ShouldAddError:        false,
		}
	}

	if err.Error() == "database role not found" {
		return ErrorHandlingResult{
			ShouldRemoveFromState: true,
			ShouldAddError:        false,
		}
	}

	return ErrorHandlingResult{
		ShouldRemoveFromState: false,
		ShouldAddError:        true,
		ErrorMessage:          "Error getting role",
	}
}

// HandlePermissionReadError analyzes an error from GetDatabasePermissionForRole and determines the appropriate action
// Note: This handles the case where the role doesn't exist (should remove entire resource)
// vs individual permissions not found (should skip that permission)
func HandlePermissionReadError(err error) ErrorHandlingResult {
	if err == nil {
		return ErrorHandlingResult{
			ShouldRemoveFromState: false,
			ShouldAddError:        false,
		}
	}

	if err.Error() == "database role not found" {
		return ErrorHandlingResult{
			ShouldRemoveFromState: true,
			ShouldAddError:        false,
		}
	}

	if err.Error() == "permissions not found" {
		// Individual permission not found - skip it, don't remove entire resource
		return ErrorHandlingResult{
			ShouldRemoveFromState: false,
			ShouldAddError:        false,
		}
	}

	return ErrorHandlingResult{
		ShouldRemoveFromState: false,
		ShouldAddError:        true,
		ErrorMessage:          "Error getting permission for role",
	}
}
