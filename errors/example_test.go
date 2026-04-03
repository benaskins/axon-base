package errors_test

import (
	"fmt"

	axerrors "github.com/benaskins/axon-base/errors"
)

func ExampleWrapError() {
	err := axerrors.WrapError("users.Create", fmt.Errorf("duplicate key"))
	fmt.Println(err)
	// Output: users.Create: duplicate key
}

func ExampleWrapError_nil() {
	err := axerrors.WrapError("users.Create", nil)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleIsNotFoundError() {
	wrapped := axerrors.WrapError("users.Get", axerrors.ErrNotFound)
	fmt.Println(axerrors.IsNotFoundError(wrapped))
	// Output: true
}

func ExampleIsNotFoundError_false() {
	err := fmt.Errorf("some other error")
	fmt.Println(axerrors.IsNotFoundError(err))
	// Output: false
}
