import React from "react";
import { useListImageAutomation } from "../../hooks/imageautomation";
import { Kind } from "../../lib/api/core/types.pb";
import { showInterval } from "../../lib/time";
import { Source } from "../../lib/types";
import DataTable from "../DataTable";
import KubeStatusIndicator from "../KubeStatusIndicator";
import LoadingWrapper from "./LoadingWrapper";

const ImageRepositoriesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    Kind.ImageRepository
  );
  const initialFilterState = {
    // ...filterConfig(versions, "version"),
  };
  return (
    <LoadingWrapper loading={isLoading} error={error}>
      <DataTable
        // className={className}
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
            value: (s: Source) => (
              <KubeStatusIndicator
                short
                conditions={s.conditions}
                suspended={s.suspended}
              />
            ),
            defaultSort: true,
          },
          {
            label: "Interval",
            value: (s: Source) => showInterval(s.interval),
          },
          {
            label: "Tag Count",
            value: "tagCount",
            // sortValue: (s: Source) => s.lastUpdatedAt || "",
          },
        ]}
      />
    </LoadingWrapper>
  );
};

export default ImageRepositoriesTable;
