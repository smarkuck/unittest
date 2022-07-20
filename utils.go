package unittest

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	setupFuncName      = "Setup"
	testPanickedFormat = "test panicked:\n%s"
	noPanicText        = " (panic didn't return error)"
	defaultMsg         = "unexpected value"

	outputTemplate = `
%s:
    actual:   %s
    expected: %s`
)

type T = testing.T

func TestSuite(t *T, suite any) {
	s := suiteGuts{
		innerType:  reflect.TypeOf(suite),
		innerValue: reflect.ValueOf(suite),
	}
	s.setupAction = s.getSetupAction()
	s.runTests(t)
}

type suiteGuts struct {
	innerType   reflect.Type
	innerValue  reflect.Value
	setupAction func()
}

func (s *suiteGuts) getSetupAction() func() {
	m := s.innerValue.MethodByName(setupFuncName)
	if m.IsValid() {
		return func() { m.Call([]reflect.Value{}) }
	}
	return func() {}
}

func (s *suiteGuts) runTests(t *T) {
	for i := 0; i < s.innerType.NumMethod(); i++ {
		method := s.innerType.Method(i)
		if method.Name != setupFuncName {
			s.runTest(t, method)
		}
	}
}

func (s *suiteGuts) runTest(t *T, test reflect.Method) {
	t.Run(test.Name, func(t *T) {
		defer recoverOnFail(t)
		s.setupAction()
		test.Func.Call([]reflect.Value{
			s.innerValue, reflect.ValueOf(t)})
	})
}

func recoverOnFail(t *T) {
	err := recover()
	if err != nil {
		t.Fatalf(testPanickedFormat, err)
	}
}

func ExpectPanicErrEq(t *T, text string, msg ...string) {
	switch err := recover().(type) {
	case error:
		ExpectEq(t, err.Error(), text, msg...)
	default:
		msg := getMsg(msg...) + noPanicText
		signalError(t, msg, "%v", err, text)
	}
}

func ExpectTrue(t *T, value bool, msg ...string) {
	ExpectEq(t, value, true, msg...)
}

func ExpectFalse(t *T, value bool, msg ...string) {
	ExpectEq(t, value, false, msg...)
}

func ExpectEq[Value comparable](t *T,
	actual, expected Value, msg ...string) {
	ExpectEqf(t, actual, expected, "%v", msg...)
}

func ExpectEqf[Value comparable](t *T,
	actual, expected Value, format string, msg ...string) {
	if actual != expected {
		signalError(t, getMsg(msg...), format,
			actual, expected)
	}
}

func ExpectDeepEq(t *T,
	actual, expected any, msg ...string) {
	if !reflect.DeepEqual(actual, expected) {
		signalError(t, getMsg(msg...), "%v, (%T)",
			actual, actual, expected, expected)
	}
}

func signalError(t *T, msg, format string, args ...any) {
	f := fmt.Sprintf(outputTemplate, msg, format, format)
	t.Errorf(f, args...)
}

func getMsg(msg ...string) string {
	if len(msg) > 0 {
		return msg[0]
	}
	return defaultMsg
}
