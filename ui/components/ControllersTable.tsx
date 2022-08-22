import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Deployment } from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";
import DataTable, { filterByStatusCallback, filterConfig } from "./DataTable";
import KubeStatusIndicator from "./KubeStatusIndicator";
import Link from "./Link";

type Props = {
  className?: string;
  controllers?: Deployment[];
};

function ControllersTable({ className, controllers = [] }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  let initialFilterState = {
    ...filterConfig(controllers, "status", filterByStatusCallback),
  };

  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(controllers, "clusterName"),
    };
  }

  return (
    <DataTable
      className={className}
      filters={initialFilterState}
      rows={controllers}
      fields={[
        {
          label: "Name",
          value: "name",
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: "Status",
          value: (d: Deployment) =>
            d.conditions.length > 0 ? (
              <KubeStatusIndicator
                short
                conditions={d.conditions}
                suspended={d.suspended}
              />
            ) : null,
          sortValue: statusSortHelper,
        },
        ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
          ? [{ label: "Cluster", value: "clusterName" }]
          : []),
        {
          value: (d: Deployment) => (
            <>
              {_.map(d.images, (img) => (
                <Link href={`https://${img}`} key={img} newTab>
                  {img}
                </Link>
              ))}
            </>
          ),
          label: "Image",
        },
      ]}
    />
  );
}

export default styled(ControllersTable).attrs({
  className: ControllersTable.name,
})``;
