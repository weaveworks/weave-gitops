import React from "react";
import { useListImageAutomation } from "../../../hooks/imageautomation";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { showInterval } from "../../../lib/time";
import { Source, V2Routes } from "../../../lib/types";
import DataTable, { filterConfig } from "../../DataTable";
import KubeStatusIndicator from "../../KubeStatusIndicator";
import Link from "../../Link";
import LoadingWrapper from "../LoadingWrapper";

const ImageRepositoriesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    Kind.ImageRepository
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, "name"),
  };
  return (
    <LoadingWrapper loading={isLoading} error={error}>
      <DataTable
        filters={initialFilterState}
        rows={data?.objects}
        fields={[
          {
            label: "Name",
            value: ({ name, namespace, clusterName }) => (
              <Link
                to={formatURL(V2Routes.ImageAutomationRepositoriesDetails, {
                  name: name,
                  namespace: namespace,
                  clusterName: clusterName,
                })}
              >
                {name}
              </Link>
            ),
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
