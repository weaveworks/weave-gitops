import React from "react";
import { useListImageAutomation } from "../../hooks/imageautomation";
import { Kind, Object } from "../../lib/api/core/types.pb";
import { Source } from "../../lib/objects";
import { showInterval } from "../../lib/time";
import DataTable from "../DataTable";
import KubeStatusIndicator from "../KubeStatusIndicator";
import SourceLink from "../SourceLink";
import Timestamp from "../Timestamp";
import LoadingWrapper from "./LoadingWrapper";

const ImageAutomationsTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    Kind.ImageUpdateAutomation
  );
  console.log(data?.objects);

  const initialFilterState = {
    // ...filterConfig(versions, "version"),
  };
  return (
    <LoadingWrapper loading={isLoading} error={error}>
      <DataTable
        filters={initialFilterState}
        rows={data?.objects}
        fields={[
          {
            label: "Name",
            value: "name",
            textSearchable: true,
            maxWidth: 600,
          },
          {
            label: "Namespace",
            value: "namespace",
          },
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
    </LoadingWrapper>
  );
};

export default ImageAutomationsTable;
