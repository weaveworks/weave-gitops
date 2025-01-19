package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	k8sFields "k8s.io/apimachinery/pkg/fields"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	sigsClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type validationList struct {
	Validations []*pb.PolicyValidation
	Token       string
	Errors      []*pb.ListError
}

const DefaultValidationType = "Admission"

func (cs *coreServer) ListPolicyValidations(ctx context.Context, m *pb.ListPolicyValidationsRequest) (*pb.ListPolicyValidationsResponse, error) {
	var respErrors []*pb.ListError

	var clustersClient clustersmngr.Client
	var err error

	if m.ClusterName != "" {
		clustersClient, err = cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	} else {
		clustersClient, err = cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	}

	if err != nil {
		var merr *multierror.Error
		if errors.As(err, &merr) {
			for _, err := range merr.Errors {
				var cerr *clustersmngr.ClientError
				if errors.As(err, &cerr) {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	var validationType string

	if m.ValidationType != "" {
		validationType = m.ValidationType
	} else {
		validationType = DefaultValidationType
	}

	labelSelector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": validationType,
	})
	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %w", err)
	}

	fieldSelectorSet := map[string]string{
		"type": "Warning",
	}

	if m.Application != "" {
		fieldSelectorSet["involvedObject.name"] = m.Application
		fieldSelectorSet["involvedObject.kind"] = m.Kind
	}

	if m.Namespace != "" {
		fieldSelectorSet["involvedObject.namespace"] = m.Namespace
	}

	if m.PolicyId != "" {
		fieldSelectorSet["involvedObject.fieldPath"] = m.PolicyId
	}

	fieldSelector := k8sFields.SelectorFromSet(fieldSelectorSet)

	opts := []sigsClient.ListOption{}
	if m.Pagination != nil {
		opts = append(opts, sigsClient.Limit(m.Pagination.PageSize))
		opts = append(opts, sigsClient.Continue(m.Pagination.PageToken))
	}
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})
	opts = append(opts, sigsClient.InNamespace(v1.NamespaceAll))

	validationsList, err := cs.listValidationsFromEvents(ctx, clustersClient, m.ClusterName, false, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}
	respErrors = append(respErrors, validationsList.Errors...)
	policyviolationlist := pb.ListPolicyValidationsResponse{
		Total:         int32(len(validationsList.Validations)),
		Violations:    validationsList.Validations,
		Errors:        respErrors,
		NextPageToken: validationsList.Token,
	}
	return &policyviolationlist, nil
}

func (cs *coreServer) GetPolicyValidation(ctx context.Context, m *pb.GetPolicyValidationRequest) (*pb.GetPolicyValidationResponse, error) {
	var clusterClient clustersmngr.Client
	var err error

	if m.ClusterName != "" {
		clusterClient, err = cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	} else {
		clusterClient, err = cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	}

	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var validationType string

	if m.ValidationType != "" {
		validationType = m.ValidationType
	} else {
		validationType = DefaultValidationType
	}

	selector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": validationType,
		"pac.weave.works/id":   m.ValidationId,
	})
	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %w", err)
	}
	opts := []sigsClient.ListOption{}

	fields := k8sFields.OneTermEqualSelector("type", "Warning")
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: selector,
		FieldSelector: fields,
	})
	opts = append(opts, sigsClient.InNamespace(v1.NamespaceAll))

	validationsList, err := cs.listValidationsFromEvents(ctx, clusterClient, m.ClusterName, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}
	if len(validationsList.Errors) > 0 {
		return nil, fmt.Errorf("error getting events: %s", validationsList.Errors[0].Message)
	}
	if len(validationsList.Validations) == 0 {
		return nil, fmt.Errorf("no policy violation found with id %s and cluster: %s", m.ValidationId, m.ClusterName)
	}
	return &pb.GetPolicyValidationResponse{
		Validation: validationsList.Validations[0],
	}, nil
}

func (cs *coreServer) listValidationsFromEvents(ctx context.Context, clusterClient clustersmngr.Client, clusterName string, extraDetails bool, opts []sigsClient.ListOption) (*validationList, error) {
	respErrors := []*pb.ListError{}
	clist := clustersmngr.NewClusteredList(func() sigsClient.ObjectList {
		return &v1.EventList{}
	})

	if err := clusterClient.ClusteredList(ctx, clist, true, opts...); err != nil {
		var errs clustersmngr.ClusteredListError
		if !errors.As(err, &errs) {
			return nil, fmt.Errorf("error while listing events: %w", err)
		}

		for _, e := range errs.Errors {
			respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
		}
	}

	var validations []*pb.PolicyValidation
	for listClusterName, lists := range clist.Lists() {
		if clusterName != "" && listClusterName != clusterName {
			continue
		}
		for _, l := range lists {
			list, ok := l.(*v1.EventList)
			if !ok {
				continue
			}
			for i := range list.Items {
				validation, err := eventToPolicyValidation(list.Items[i], listClusterName, extraDetails)
				if err != nil {
					return nil, fmt.Errorf("error while getting policy violation event details: %w", err)
				}
				validations = append(validations, validation)
			}
		}
	}

	return &validationList{
		Validations: validations,
		Token:       clist.GetContinue(),
		Errors:      respErrors,
	}, nil
}

func eventToPolicyValidation(item v1.Event, clusterName string, extraDetails bool) (*pb.PolicyValidation, error) {
	annotations := item.GetAnnotations()
	policyValidation := &pb.PolicyValidation{
		Id:          ExtractStringValueFromMap(item.GetLabels(), "pac.weave.works/id"),
		Name:        ExtractStringValueFromMap(annotations, "policy_name"),
		PolicyId:    ExtractStringValueFromMap(annotations, "policy_id"),
		ClusterId:   ExtractStringValueFromMap(annotations, "cluster_id"),
		Category:    ExtractStringValueFromMap(annotations, "category"),
		Severity:    ExtractStringValueFromMap(annotations, "severity"),
		CreatedAt:   item.GetCreationTimestamp().Format(time.RFC3339),
		Message:     item.Message,
		Entity:      item.InvolvedObject.Name,
		EntityKind:  item.InvolvedObject.Kind,
		Namespace:   item.InvolvedObject.Namespace,
		ClusterName: clusterName,
	}
	if extraDetails {
		policyValidation.Description = ExtractStringValueFromMap(annotations, "description")
		policyValidation.HowToSolve = ExtractStringValueFromMap(annotations, "how_to_solve")
		policyValidation.ViolatingEntity = ExtractStringValueFromMap(annotations, "entity_manifest")
		err := json.Unmarshal([]byte(ExtractStringValueFromMap(annotations, "occurrences")), &policyValidation.Occurrences)
		if err != nil {
			return nil, fmt.Errorf("failed to get occurrences from event: %w", err)
		}
		paramsRaw := ExtractStringValueFromMap(annotations, "parameters")
		if paramsRaw != "" {
			parameter, err := getPolicyValidationParam([]byte(paramsRaw))
			if err != nil {
				return nil, err
			}
			policyValidation.Parameters = parameter
		}
	}
	return policyValidation, nil
}

func getPolicyValidationParam(raw []byte) ([]*pb.PolicyValidationParam, error) {
	var paramsArr []map[string]interface{}
	err := json.Unmarshal(raw, &paramsArr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy validation parameter, error: %w", err)
	}

	parameters := make([]*pb.PolicyValidationParam, len(paramsArr))
	for i := range paramsArr {
		param := pb.PolicyValidationParam{
			Name:     paramsArr[i]["name"].(string),
			Type:     paramsArr[i]["type"].(string),
			Required: paramsArr[i]["required"].(bool),
		}

		if val, ok := paramsArr[i]["config_ref"]; ok {
			param.ConfigRef = val.(string)
		}

		val, err := getParamValue(paramsArr[i]["value"])
		if err != nil {
			return nil, err
		}
		param.Value = val
		parameters[i] = &param
	}
	return parameters, nil
}

func getParamValue(param interface{}) (*anypb.Any, error) {
	if param == nil {
		return nil, nil
	}
	switch val := param.(type) {
	case string:
		value := wrapperspb.String(val)
		return anypb.New(value)
	case float64:
		value := wrapperspb.Double(val)
		return anypb.New(value)
	case bool:
		value := wrapperspb.Bool(val)
		return anypb.New(value)
	case []interface{}:
		var values []string
		for i := range val {
			values = append(values, fmt.Sprintf("%v", val[i]))
		}
		value := &pb.PolicyParamRepeatedString{Value: values}
		return anypb.New(value)
	}
	return nil, nil
}
