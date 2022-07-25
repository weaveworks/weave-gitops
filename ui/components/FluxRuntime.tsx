import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { SortType } from "../components/DataTable";
import Flex from "../components/Flex";
import KubeStatusIndicator from "../components/KubeStatusIndicator";
import Link from "../components/Link";
import Text from "../components/Text";
import { Crd, Deployment } from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";
import { filterByStatusCallback, filterConfig } from "./FilterableTable";
import URLAddressableTable from "./URLAddressableTable";

type Props = {
  className?: string;
  deployments?: Deployment[];
  crds?: Crd[];
};

function FluxRuntime({ className, deployments, crds }: Props) {
  const initialFilterState = {
    ...filterConfig(deployments, "clusterName"),
    ...filterConfig(deployments, "status", filterByStatusCallback),
  };

  const crdFilterState = {
    ...filterConfig(crds, "version"),
    ...filterConfig(crds, "kind"),
    ...filterConfig(crds, "clusterName"),
  };

  return (
    <Flex wide tall column>
      <Text size="large" semiBold titleHeight>
        Controllers
      </Text>
      <URLAddressableTable
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
      />
      <Text size="large" semiBold titleHeight>
        Custom Resource Definitions
      </Text>
      <URLAddressableTable
        className={className}
        filters={crdFilterState}
        rows={crds}
        fields={[
          {
            label: "Name",
            value: "name",
            textSearchable: true,
            maxWidth: 600,
          },
          {
            label: "Kind",
            value: "kind",
          },
          {
            label: "Version",
            value: "version",
          },
          {
            label: "Cluster",
            value: "clusterName",
          },
        ]}
      />
    </Flex>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
