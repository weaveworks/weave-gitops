import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import DataTable, { SortType } from "../components/DataTable";
import KubeStatusIndicator from "../components/KubeStatusIndicator";
import Link from "../components/Link";
import { Deployment } from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";

type Props = {
  className?: string;
  deployments?: Deployment[];
};

function FluxRuntime(props: Props) {
  return (
    <DataTable
      className={props.className}
      defaultSort={2}
      fields={[
        {
          label: "Name",
          value: "name",
        },
        {
          value: (v: Deployment) => (
            <KubeStatusIndicator
              conditions={v.conditions}
              suspended={v.suspended}
            />
          ),
          label: "Status",
          sortValue: statusSortHelper,
          sortType: SortType.number,
        },
        {
          label: "Cluster",
          value: "clusterName",
        },
        {
          value: (v: Deployment) => (
            <>
              {_.map(v.images, (img) => (
                <Link href={`https://${img}`} key={img} newTab>
                  {img}
                </Link>
              ))}
            </>
          ),
          label: "Image",
        },
      ]}
      rows={props.deployments}
    />
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
