package testutils

import (
	"fmt"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/onsi/gomega/types"
)

func MatchGRPCError(code interface{}, err interface{}) types.GomegaMatcher {
	return &grpcErrorMatcher{
		codeExpected: code,
		errExpected:  err,
	}
}

type grpcErrorMatcher struct {
	codeExpected interface{}
	errExpected  interface{}
}

func (matcher *grpcErrorMatcher) Match(actual interface{}) (success bool, err error) {
	actualErr, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("MatchGRPCError matcher expects an error")
	}

	status, ok := status.FromError(actualErr)
	if !ok {
		return false, fmt.Errorf("MatchGRPCError cannot convert error to status.Status")
	}

	codeExpected := matcher.codeExpected.(codes.Code)
	errExpected := matcher.errExpected.(error)

	return status.Code() == codeExpected && status.Message() == errExpected.Error(), nil
}

func (matcher *grpcErrorMatcher) FailureMessage(actual interface{}) (message string) {
	actualStatus, _ := status.FromError(actual.(error))

	return fmt.Sprintf("Expected \n\t%s\n to match: \n\t%s\nAnd \n\t%v\n to match: \n\t%v\n", actualStatus.Code().String(), matcher.codeExpected.(codes.Code).String(), actualStatus.Message(), matcher.errExpected.(error).Error())
}

func (matcher *grpcErrorMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualStatus, _ := status.FromError(actual.(error))

	return fmt.Sprintf("Expected \n\t%s\n not to match: \n\t%s\nAnd \n\t%v\n not to match: \n\t%v\n", actualStatus.Code().String(), matcher.codeExpected.(codes.Code).String(), actualStatus.Message(), matcher.errExpected.(error).Error())
}
