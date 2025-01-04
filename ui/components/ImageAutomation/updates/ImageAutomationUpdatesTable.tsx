import React from "react";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { useListImageAutomation } from "../../../hooks/imageautomation";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { showInterval } from "../../../lib/time";
import { V2Routes } from "../../../lib/types";
import DataTable, { filterConfig } from "../../DataTable";
import KubeStatusIndicator from "../../KubeStatusIndicator";
import Link from "../../Link";
import RequestStateHandler from "../../RequestStateHandler";
import SourceLink from "../../SourceLink";
import Timestamp from "../../Timestamp";

const ImageAutomationUpdatesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    Kind.ImageUpdateAutomation,
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, "name"),
  };
  const { isFlagEnabled } = useFeatureFlags();
  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <DataTable
        hasCheckboxes
        filters={initialFilterState}
        rows={data?.objects}
        fields={[
          {
            label: "Name",
            value: ({ name, namespace, clusterName }) => (
              <Link
                to={formatURL(V2Routes.ImageAutomationUpdatesDetails, {
                  name: name,
                  namespace: namespace,
                  clusterName: clusterName,
                })}
              >
                {name}
              </Link>
            ),
            textSearchable: true,
            sortValue: ({ name }) => name || "",
            maxWidth: 600,
          },
          {
            label: "Namespace",
            value: "namespace",
          },
          ...(isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
            ? [
                {
                  label: "Cluster",
                  value: "clusterName",
                  sortValue: ({ clusterName }) => clusterName,
                },
              ]
            : []),
          {
            label: "Status",
            value: ({ conditions, suspended }) => (
              <KubeStatusIndicator
                short
                conditions={conditions}
                suspended={suspended}
              />
            ),
            defaultSort: true,
          },
          {
            label: "Source",
            value: ({ sourceRef, clusterName }) => (
              <SourceLink sourceRef={sourceRef} clusterName={clusterName} />
            ),
          },
          {
            label: "Interval",
            value: ({ interval }) => showInterval(interval),
          },
          {
            label: "Last Run",
            value: ({ lastAutomationRunTime }) => (
              <Timestamp time={lastAutomationRunTime} />
            ),
          },
        ]}
      />
    </RequestStateHandler>
  );
};

export default ImageAutomationUpdatesTable;
