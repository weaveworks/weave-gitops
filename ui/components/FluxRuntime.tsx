import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { SortType } from "../components/DataTable";
import KubeStatusIndicator from "../components/KubeStatusIndicator";
import Link from "../components/Link";
import { Deployment } from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";
import FilterableTable, {
  filterConfig,
  filterByStatusCallback,
} from "./FilterableTable";

type Props = {
  className?: string;
  deployments?: Deployment[];
};

function FluxRuntime({ className, deployments }: Props) {
  const initialFilterState = {
    ...filterConfig(deployments, "clusterName"),
    ...filterConfig(deployments, "status", filterByStatusCallback),
  };

  return (
    <FilterableTable
      className={className}
      filters={initialFilterState}
      rows={deployments}
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
          sortType: SortType.number,
          sortValue: statusSortHelper,
        },
        {
          label: "Cluster",
          value: "clusterName",
        },
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
    ></FilterableTable>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
