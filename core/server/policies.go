package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errRequiredClusterName = errors.New("`clusterName` param is required")

func getPolicyParamValue(param pacv2beta2.PolicyParameters, policyID string) (*anypb.Any, error) {
	if param.Value == nil {
		return nil, nil
	}
	var anyValue *any.Any
	var err error
	switch param.Type {
	case "string":
		var strValue string
		// attempt to clean up extra quotes if not successful show as is
		unquotedValue, UnquoteErr := strconv.Unquote(string(param.Value.Raw))
		if UnquoteErr != nil {
			strValue = string(param.Value.Raw)
		} else {
			strValue = unquotedValue
		}
		value := wrapperspb.String(strValue)
		anyValue, err = anypb.New(value)
	case "integer":
		intValue, convErr := strconv.Atoi(string(param.Value.Raw))
		if convErr != nil {
			err = convErr
			break
		}
		value := wrapperspb.Int32(int32(intValue))
		anyValue, err = anypb.New(value)
	case "boolean":
		boolValue, convErr := strconv.ParseBool(string(param.Value.Raw))
		if convErr != nil {
			err = convErr
			break
		}
		value := wrapperspb.Bool(boolValue)
		anyValue, err = anypb.New(value)
	case "array":
		var arrayValue []string
		convErr := json.Unmarshal(param.Value.Raw, &arrayValue)
		if convErr != nil {
			err = convErr
			break
		}
		value := &pb.PolicyParamRepeatedString{Value: arrayValue}
		anyValue, err = anypb.New(value)
	default:
		return nil, fmt.Errorf("found unsupported policy parameter type %s in policy %s", param.Type, policyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to serialize parameter value %s in policy %s: %w", param.Name, policyID, err)
	}
	return anyValue, nil
}

func toPolicyResponse(policyCRD pacv2beta2.Policy, clusterName string) (*pb.Policy, error) {
	policySpec := policyCRD.Spec

	var policyLabels []*pb.PolicyTargetLabel
	for i := range policySpec.Targets.Labels {
		policyLabels = append(policyLabels, &pb.PolicyTargetLabel{
			Values: policySpec.Targets.Labels[i],
		})
	}

	var policyParams []*pb.PolicyParam
	for _, param := range policySpec.Parameters {
		policyParam := &pb.PolicyParam{
			Name:     param.Name,
			Required: param.Required,
			Type:     param.Type,
		}
		value, err := getPolicyParamValue(param, policySpec.ID)
		if err != nil {
			return nil, err
		}
		policyParam.Value = value
		policyParams = append(policyParams, policyParam)
	}
	var policyStandards []*pb.PolicyStandard
	for _, standard := range policySpec.Standards {
		policyStandards = append(policyStandards, &pb.PolicyStandard{
			Id:       standard.ID,
			Controls: standard.Controls,
		})
	}
	policy := &pb.Policy{
		Name:        policySpec.Name,
		Id:          policySpec.ID,
		Code:        policySpec.Code,
		Description: policySpec.Description,
		HowToSolve:  policySpec.HowToSolve,
		Category:    policySpec.Category,
		Tags:        policySpec.Tags,
		Severity:    policySpec.Severity,
		Standards:   policyStandards,
		Targets: &pb.PolicyTargets{
			Kinds:      policySpec.Targets.Kinds,
			Namespaces: policySpec.Targets.Namespaces,
			Labels:     policyLabels,
		},
		Parameters:  policyParams,
		CreatedAt:   policyCRD.CreationTimestamp.Format(time.RFC3339),
		ClusterName: clusterName,
		Tenant:      policyCRD.GetLabels()["toolkit.fluxcd.io/tenant"],
		Modes:       policyCRD.Status.Modes,
	}

	return policy, nil
}

func (cs *coreServer) ListPolicies(ctx context.Context, m *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	respErrors := []*pb.ListError{}
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
		}
	}

	opts := []client.ListOption{}
	if m.Pagination != nil {
		opts = append(opts, client.Limit(m.Pagination.PageSize))
		opts = append(opts, client.Continue(m.Pagination.PageToken))
	}

	var continueToken string
	var listsV2beta2 map[string][]client.ObjectList
	var listsV2beta1 map[string][]client.ObjectList

	if m.ClusterName == "" {
		clistV2beta2 := clustersmngr.NewClusteredList(func() client.ObjectList {
			return &pacv2beta2.PolicyList{}
		})
		clistV2beta1 := clustersmngr.NewClusteredList(func() client.ObjectList {
			return &pacv2beta1.PolicyList{}
		})

		var errsV2beta2 clustersmngr.ClusteredListError
		var errsV2beta1 clustersmngr.ClusteredListError

		if err := clustersClient.ClusteredList(ctx, clistV2beta2, false, opts...); err != nil {
			if !errors.As(err, &errsV2beta2) {
				return nil, fmt.Errorf("error while listing v2beta2 policies: %w", err)
			}
		}
		for _, e := range errsV2beta2.Errors {
			if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
				respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
			}
		}

		if err := clustersClient.ClusteredList(ctx, clistV2beta1, false, opts...); err != nil {
			if !errors.As(err, &errsV2beta1) {
				return nil, fmt.Errorf("error while listing v2beta1 policies: %w", err)
			}
		}
		for _, e := range errsV2beta1.Errors {
			if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
				respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
			}
		}

		continueToken = clistV2beta2.GetContinue()
		listsV2beta2 = clistV2beta2.Lists()
		listsV2beta1 = clistV2beta1.Lists()
	} else {
		listV2beta2 := &pacv2beta2.PolicyList{}
		listV2beta1 := &pacv2beta1.PolicyList{}

		policiesV2beta2, policiesV2beta1 := true, true

		if err := clustersClient.List(ctx, m.ClusterName, listV2beta2, opts...); err != nil {
			policiesV2beta2 = false
		}
		if err := clustersClient.List(ctx, m.ClusterName, listV2beta1, opts...); err != nil {
			policiesV2beta1 = false
		}

		if !(policiesV2beta2 || policiesV2beta1) {
			return nil, fmt.Errorf("error while listing policies for cluster %s: %w", m.ClusterName, err)
		}

		continueToken = listV2beta2.GetContinue()

		if policiesV2beta1 {
			listsV2beta1 = map[string][]client.ObjectList{m.ClusterName: {listV2beta1}}
		}
		if policiesV2beta2 {
			listsV2beta2 = map[string][]client.ObjectList{m.ClusterName: {listV2beta2}}
		}
	}

	var policies []*pb.Policy
	collectedPolicies := map[string]struct{}{}
	for clusterName, lists := range listsV2beta2 {
		for _, l := range lists {
			list, ok := l.(*pacv2beta2.PolicyList)
			if !ok {
				continue
			}
			for i := range list.Items {
				policy, err := toPolicyResponse(list.Items[i], clusterName)
				if err != nil {
					return nil, err
				}

				policies = append(policies, policy)
				collectedPolicies[getClusterPolicyKey(clusterName, list.Items[i].GetName())] = struct{}{}
			}
		}
	}
	for clusterName, lists := range listsV2beta1 {
		for _, l := range lists {
			list, ok := l.(*pacv2beta1.PolicyList)
			if !ok {
				continue
			}
			for i := range list.Items {
				if _, ok := collectedPolicies[getClusterPolicyKey(clusterName, list.Items[i].GetName())]; ok {
					continue
				}
				policy, err := toPolicyResponseV2beta1(list.Items[i], clusterName)
				if err != nil {
					return nil, err
				}
				policies = append(policies, policy)
			}
		}
	}

	return &pb.ListPoliciesResponse{
		Policies:      policies,
		Total:         int32(len(policies)),
		NextPageToken: continueToken,
		Errors:        respErrors,
	}, nil
}

func (cs *coreServer) GetPolicy(ctx context.Context, m *pb.GetPolicyRequest) (*pb.GetPolicyResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if m.ClusterName == "" {
		return nil, errRequiredClusterName
	}
	policyCRv2beta2 := pacv2beta2.Policy{}
	policyCRv2beta1 := pacv2beta1.Policy{}
	policiesV2beta2, policiesV2beta1 := true, true
	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCRv2beta2); err != nil {
		policiesV2beta2 = false
	}
	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCRv2beta1); err != nil {
		policiesV2beta1 = false
	}
	if !(policiesV2beta2 || policiesV2beta1) {
		return nil, fmt.Errorf("error while getting policy %s from cluster %s: %w", m.PolicyName, m.ClusterName, err)
	}

	var policy *pb.Policy
	if policiesV2beta1 {
		policy, err = toPolicyResponseV2beta1(policyCRv2beta1, m.ClusterName)
		if err != nil {
			return nil, err
		}
	}
	if policiesV2beta2 {
		policy, err = toPolicyResponse(policyCRv2beta2, m.ClusterName)
		if err != nil {
			return nil, err
		}
	}

	return &pb.GetPolicyResponse{Policy: policy}, nil
}

func getClusterPolicyKey(clusterName, policyId string) string {
	return fmt.Sprintf("%s.%s", clusterName, policyId)
}
