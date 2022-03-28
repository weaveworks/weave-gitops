import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import DataTable, { SortType } from "../../components/DataTable";
import KubeStatusIndicator, {
  computeReady,
} from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
import Spacer from "../../components/Spacer";
import { useListFluxRuntimeObjects } from "../../hooks/flux";
import { Deployment } from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
};

function FluxRuntime({ className }: Props) {
  const { data, isLoading, error } = useListFluxRuntimeObjects();

  return (
    <Page
      title="Flux Runtime"
      loading={isLoading}
      error={error}
      className={className}
    >
      <Spacer padding="xs" />
      <DataTable
        defaultSort={2}
        fields={[
          { value: "name", label: "Name" },
          {
            value: (v: Deployment) => (
              <KubeStatusIndicator
                conditions={v.conditions}
                suspended={v.suspended}
              />
            ),
            label: "Status",
            sortType: SortType.bool,
            sortValue: ({ conditions }) => computeReady(conditions),
          },
          {
            label: "Cluster",
            value: () => "Default",
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
        rows={data?.deployments}
      />
    </Page>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
