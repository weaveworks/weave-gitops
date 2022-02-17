import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import DataTable from "../../components/DataTable";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
import { useListFluxRuntimeObjects } from "../../hooks/apps";
import { Deployment } from "../../lib/api/app/apps.pb";

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
      <DataTable
        sortFields={["cluster"]}
        fields={[
          { value: "name", label: "Name" },
          {
            value: (v: Deployment) => (
              <KubeStatusIndicator conditions={v.conditions} />
            ),
            label: "Status",
          },
          {
            label: "Cluster",
            value: "cluster",
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
