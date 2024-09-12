package provider

import "fmt"

type listNotEmpty struct{}

func (listNotEmpty) CheckValue(v any) error {
	val, ok := v.([]any)

	if !ok {
		return fmt.Errorf("expected []any value for ListNotEmpty check, got: %T", v)
	}

	if len(val) == 0 {
		return fmt.Errorf("expected non-empty list for ListNotEmpty check, but list was empty")
	}

	return nil
}

func (listNotEmpty) String() string {
	return "non-empty list"
}

type listOfNonNulls struct{}

func (listOfNonNulls) CheckValue(v any) error {
	val, ok := v.([]any)

	if !ok {
		return fmt.Errorf("expected []any value for ListOfNonNulls check, got: %T", v)
	}

	for i, item := range val {
		if item == nil {
			return fmt.Errorf("expected non-nil value at index %d for ListOfNonNulls check, but value was nil", i)
		}
	}

	return nil
}

func (listOfNonNulls) String() string {
	return "list of non-nil values"
}
