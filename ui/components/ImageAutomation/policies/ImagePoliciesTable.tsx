import React from "react";
import { useListImageAutomation } from "../../../hooks/imageautomation";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImgPolicy } from "../../../lib/objects";
import { Source } from "../../../lib/types";
import DataTable, { filterConfig } from "../../DataTable";
import KubeStatusIndicator from "../../KubeStatusIndicator";
import RequestStateHandler from "../../RequestStateHandler";

const ImagePoliciesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(Kind.ImagePolicy);
  const initialFilterState = {
    ...filterConfig(data?.objects, "name"),
    ...filterConfig(data?.objects, "imageRepositoryRef"),
  };
  return (
    <RequestStateHandler loading={isLoading} error={error}>
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
            label: "Image Policy",
            value: ({ imagePolicy }: { imagePolicy: ImgPolicy }) =>
              imagePolicy?.type || "",
          },
          {
            label: "Order/Range",
            value: ({ imagePolicy }: { imagePolicy: ImgPolicy }) =>
              imagePolicy?.value || "",
          },
          {
            label: "Image Repository",
            value: "imageRepositoryRef",
          },
        ]}
      />
    </RequestStateHandler>
  );
};

export default ImagePoliciesTable;
