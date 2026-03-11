package response

const OK = "OK"
const SentResetToken = "Please check your email for your 6-digit PIN."
const Created = "created"
const Updated = "updated"
const Deleted = "deleted"
const Unauthorized = "unauthorized"
const UnauthorizedSystem = "unauthorized system"
const ServerError = "server error"
const ModelNotFound = "model not found"
const ValidationError = "validation error"

func SuccessResponse(data interface{}, message string) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}
}

func ErrorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
		"data":    nil,
	}
}

func ErrorResponseWithData(message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
		"data":    data,
	}
}

func ValidationErrorResponse(message interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
		"data":    nil,
	}
}
