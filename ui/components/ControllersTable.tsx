import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Deployment } from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";
import { SortType } from "./DataTable";
import { filterByStatusCallback, filterConfig } from "./FilterableTable";
import KubeStatusIndicator from "./KubeStatusIndicator";
import Link from "./Link";
import URLAddressableTable from "./URLAddressableTable";

type Props = {
  className?: string;
  controllers?: Deployment[];
};

function ControllersTable({ className, controllers = [] }: Props) {
  const initialFilterState = {
    ...filterConfig(controllers, "clusterName"),
    ...filterConfig(controllers, "status", filterByStatusCallback),
  };
  return (
    <URLAddressableTable
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
    />
  );
}

export default styled(ControllersTable).attrs({
  className: ControllersTable.name,
})``;
