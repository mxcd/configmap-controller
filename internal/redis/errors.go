package redis

type BusyGroupError struct{}

func (e *BusyGroupError) Error() string {
	return "BUSYGROUP Consumer Group name already exists"
}

type NoDataError struct{}

func (e *NoDataError) Error() string {
	return "redis: nil"
}

func ParseError(err error) error {
	if err == nil {
		return nil
	}
	if err.Error() == "BUSYGROUP Consumer Group name already exists" {
		return &BusyGroupError{}
	} else if err.Error() == "redis: nil" {
		return &NoDataError{}
	}
	return err
}
