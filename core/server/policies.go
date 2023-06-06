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
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultCluster = "Default"
)

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

func policyToPolicyRespone(policyCRD pacv2beta2.Policy, clusterName string, fullDetails bool) (*pb.PolicyObj, error) {
	policySpec := policyCRD.Spec

	policy := &pb.PolicyObj{
		Name:      policySpec.Name,
		Id:        policySpec.ID,
		Category:  policySpec.Category,
		Tags:      policySpec.Tags,
		Severity:  policySpec.Severity,
		CreatedAt: policyCRD.CreationTimestamp.Format(time.RFC3339),
		Tenant:    policyCRD.GetLabels()["toolkit.fluxcd.io/tenant"],
		Modes:     policyCRD.Status.Modes,
	}

	if fullDetails {
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

		policy.Code = policySpec.Code
		policy.Description = policySpec.Description
		policy.HowToSolve = policySpec.HowToSolve
		policy.Standards = policyStandards
		policy.Targets = &pb.PolicyTargets{
			Kinds:      policySpec.Targets.Kinds,
			Namespaces: policySpec.Targets.Namespaces,
			Labels:     policyLabels,
		}
		policy.Parameters = policyParams
		policy.ClusterName = clusterName
	}

	return policy, nil
}

func (cs *coreServer) ListPolicies(ctx context.Context, m *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	respErrors := []*pb.ListError{}

	var clustersClient clustersmngr.Client
	var err error

	clustersClient, err = cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
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
	var lists map[string][]client.ObjectList

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &pacv2beta2.PolicyList{}
	})

	var errsV2beta2 clustersmngr.ClusteredListError

	if err := clustersClient.ClusteredList(ctx, clist, false, opts...); err != nil {
		if !errors.As(err, &errsV2beta2) {
			return nil, fmt.Errorf("error while listing v2beta2 policies: %w", err)
		}
	}
	for _, e := range errsV2beta2.Errors {
		if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
			respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
		}
	}

	continueToken = clist.GetContinue()
	lists = clist.Lists()

	var policies []*pb.PolicyObj
	for clusterName, lists := range lists {
		for _, l := range lists {
			list, ok := l.(*pacv2beta2.PolicyList)
			if !ok {
				respErrors = append(respErrors, &pb.ListError{ClusterName: clusterName, Message: fmt.Sprintf("unexpected list type %T", l)})
				continue
			}
			for i := range list.Items {
				policy, err := policyToPolicyRespone(list.Items[i], clusterName, false)
				if err != nil {
					return nil, fmt.Errorf("error while converting policy %s to response: %w", list.Items[i].Name, err)
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
	var clustersClient clustersmngr.Client
	var err error

	if m.ClusterName == "" {
		m.ClusterName = DefaultCluster
	}

	clustersClient, err = cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	policyCR := pacv2beta2.Policy{}

	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCR); err != nil {
		return nil, fmt.Errorf("error while getting policy %s from cluster %s: %w", m.PolicyName, m.ClusterName, err)
	}

	var policy *pb.PolicyObj

	policy, err = policyToPolicyRespone(policyCR, m.ClusterName, true)
	if err != nil {
		return nil, err
	}

	return &pb.GetPolicyResponse{Policy: policy}, nil
}
