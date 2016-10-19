package common

const DATE_FORMAT_SQL = "2006-01-02 15:04:05.999999999-07:00"

func IfOrInt(val bool, trueVal, falseVal int) int {
	if val {
		return trueVal
	}
	return falseVal
}
