package main

import (
	"errors"
	"fmt"
)

var (
	errResourceExists    = errors.New("resource already exists")
	errResourceNotFound  = errors.New("resource not found")
	errResourceWrongType = errors.New("resource is not of the expected type")
)

func errorResourceExists(resourceName string) error {
	return fmt.Errorf("resource %q: %w", resourceName, errResourceExists)
}

func errorResourceNotFound(resourceName string) error {
	return fmt.Errorf("resource %q: %w", resourceName, errResourceNotFound)
}

func errorResourceWrongType(resourceName string, expectedType string) error {
	return fmt.Errorf("resource %q (expected %q): %w", resourceName, expectedType, errResourceWrongType)
}
