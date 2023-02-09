import React from "react";
import { useListImageAutomation } from "../../../hooks/imageautomation";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { ImgPolicy } from "../../../lib/objects";
import { Source, V2Routes } from "../../../lib/types";
import DataTable, { filterConfig } from "../../DataTable";
import KubeStatusIndicator from "../../KubeStatusIndicator";
import Link from "../../Link";
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
            value: ({ name, namespace, clusterName }) => (
              <Link
                to={formatURL(V2Routes.ImagePolicyDetails, {
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
            value: ({ imageRepositoryRef, namespace, clusterName }) => (
              <Link
                to={formatURL(V2Routes.ImageAutomationRepositoryDetails, {
                  name: imageRepositoryRef,
                  namespace: namespace,
                  clusterName: clusterName,
                })}
              >
                {imageRepositoryRef}
              </Link>
            ),
          },
        ]}
      />
    </RequestStateHandler>
  );
};

export default ImagePoliciesTable;
